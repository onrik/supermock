package app

import (
	"fmt"
	"log/slog"

	"github.com/onrik/supermock/pkg/db"
	"github.com/onrik/supermock/pkg/handlers"

	"github.com/labstack/echo/v4"
)

func Start(addr, dbDSN string) error {
	db, err := db.New(dbDSN)
	if err != nil {
		return fmt.Errorf("connect to db error: %w", err)
	}

	defer db.Close()

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

	slog.Info(fmt.Sprintf("Listen http://%s ...", addr))
	if err := server.Start(addr); err != nil {
		return err
	}
	return nil
}
