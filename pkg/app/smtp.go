package app

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/mail"
	"strconv"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/onrik/supermock/pkg/models"
)

func parseEmail(s string) (models.Email, error) {
	email := Email{
		Raw: s,
	}
	message, err := mail.ReadMessage(bytes.NewBufferString(s))
	if err != nil {
		return email, err
	}

	email.From = message.Header.Get("From")
	email.To = message.Header.Get("To")
	email.Date = message.Header.Get("Date")
	email.Subject = message.Header.Get("Subject")
	email.ContentType = message.Header.Get("Content-Type")

	body, err := io.ReadAll(message.Body)
	if err != nil {
		return email, err
	}

	email.Body = string(body)

	return email, nil
}

type SMTP struct {
	addr   string
	server *smtpmock.Server
}

func newSMTP(addr string) *SMTP {
	if addr == "" {
		return nil
	}

	return &SMTP{
		addr: addr,
	}
}

func (s *SMTP) Start() error {
	host, portStr, err := net.SplitHostPort(s.addr)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	s.server = smtpmock.New(smtpmock.ConfigurationAttr{
		HostAddress: host,
		PortNumber:  port,
	})

	slog.Info(fmt.Sprintf("Listen smtp://%s ...", s.addr))
	return s.server.Start()
}

func (s *SMTP) Stop() error {
	return s.server.Stop()
}

func (s *SMTP) Emails() []models.Email {
	emails := []models.Email{}
	for _, m := range s.server.Messages() {
		email, err := parseEmail(m.MsgRequest())
		if err != nil {
			slog.Error("Parse email error", "error", err)
		}

		emails = append(emails, email)
	}

	return emails
}

func (s *SMTP) Purge() {
	s.server.MessagesAndPurge()
}
