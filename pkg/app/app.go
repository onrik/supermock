package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/onrik/supermock/pkg/db"
	"github.com/onrik/supermock/pkg/handlers"
	"github.com/onrik/supermock/pkg/models"

	"github.com/labstack/echo/v4"
)

type Response = models.Response
type Request = models.Request
type Email = models.Email

type Supermock struct {
	httpAddr string
	db       *db.DB
	server   *echo.Echo
	smtp     *SMTP
}

func New(httpAddr, dbDSN, smtpAddr string) (*Supermock, error) {
	db, err := db.New(dbDSN)
	if err != nil {
		return nil, fmt.Errorf("connect to db error: %w", err)
	}

	smtp := newSMTP(smtpAddr)
	h := handlers.New(db, smtp)

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.Validator = Validator{}

	server.POST("/_responses", h.ResponseCreate)
	server.GET("/_responses", h.ResponseList)
	server.DELETE("/_responses/:uuid", h.DeleteResponse)
	server.GET("/_requests/:test_id", h.Requests)
	server.GET("/_requests", h.Requests)
	server.DELETE("/_tests/:test_id", h.Clean)
	server.Any("/*", h.Catch)

	if smtp != nil {
		server.GET("/_emails", h.Emails)
		server.DELETE("/_emails", h.EmailsDelete)
	}

	return &Supermock{
		httpAddr: httpAddr,
		db:       db,
		server:   server,
		smtp:     smtp,
	}, nil
}

func (s *Supermock) Start() {
	go func() {
		if err := s.Run(); err != nil {
			panic(err)
		}
	}()
}

func (s *Supermock) Run() error {
	if s.smtp != nil {
		if err := s.smtp.Start(); err != nil {
			return err
		}
	}

	slog.Info(fmt.Sprintf("Listen http://%s ...", s.httpAddr))
	if err := s.server.Start(s.httpAddr); err != nil {
		return err
	}
	return nil
}

func (s *Supermock) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.smtp != nil {
		err := s.smtp.Stop()
		if err != nil {
			slog.Error(err.Error())
		}
	}

	err := s.server.Shutdown(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	s.db.Close()
}

func (s *Supermock) Put(ctx context.Context, responses ...Response) error {
	return s.db.ResponsesSave(ctx, responses...)
}

func (s *Supermock) Get(ctx context.Context, testID string) ([]Request, error) {
	return s.db.Requests(ctx, testID)
}
