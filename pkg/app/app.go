package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/onrik/supermock/pkg/db"
	"github.com/onrik/supermock/pkg/handlers"

	"github.com/labstack/echo/v4"
)

type Supermock struct {
	addr   string
	db     *db.DB
	server *echo.Echo
}

func New(addr, dbDSN string) (*Supermock, error) {
	db, err := db.New(dbDSN)
	if err != nil {
		return nil, fmt.Errorf("connect to db error: %w", err)
	}

	h := handlers.New(db)

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.Validator = Validator{}

	server.POST("/_responses", h.Responses)
	server.DELETE("/_responses/:uuid", h.DeleteResponse)
	server.GET("/_requests/:test_id", h.Requests)
	server.DELETE("/_tests/:test_id", h.Clean)
	server.Any("/*", h.Catch)

	return &Supermock{
		addr:   addr,
		db:     db,
		server: server,
	}, nil
}

func (s *Supermock) Start() error {
	slog.Info(fmt.Sprintf("Listen http://%s ...", s.addr))
	if err := s.server.Start(s.addr); err != nil {
		return err
	}
	return nil
}

func (s *Supermock) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	s.db.Close()
}
