package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	echoslog "github.com/onrik/echo-slog"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8000", "")
	dbPath := flag.String("db", "sqlite://db.sqlite3", "")
	debug := flag.Bool("debug", false, "")
	flag.Parse()

	logLevel := slog.LevelInfo
	if debug != nil && *debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.SourceKey {
				return a
			}

			source, ok := a.Value.Any().(*slog.Source)
			if !ok {
				return a
			}
			parts := strings.Split(source.File, "/")
			var f string
			if len(parts) >= 4 {
				f = strings.Join(parts[len(parts)-2:], "/")
			} else {
				f = strings.Join(parts, "/")
			}
			return slog.Any(
				a.Key,
				fmt.Sprintf("%s:%d", f, source.Line),
			)
		},
	})))

	db, err := NewDB(*dbPath)
	if err != nil {
		slog.Error("Open db error", "error", err)
		return
	}

	defer db.Close()

	h := Handlers{db}

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.Validator = Validator{}
	server.Use(echoslog.MiddlewareDefault())

	server.POST("/_responses", h.Responses)
	server.DELETE("/_responses/:uuid", h.DeleteResponse)
	server.GET("/_requests/:test_id", h.Requests)
	server.DELETE("/_tests/:test_id", h.Clean)
	server.Any("/*", h.Catch)

	slog.Info(fmt.Sprintf("Listen http://%s ...", *listen), "db", *dbPath)
	if err := server.Start(*listen); err != nil {
		slog.Error(err.Error())
	}
}
