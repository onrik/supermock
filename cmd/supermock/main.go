package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/onrik/supermock/pkg/app"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	config, err := readConfig()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     config.slogLevel(),
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

	supermock, err := app.New(config.HttpAddr, config.DB, config.SmtpAddr)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if err := supermock.Run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
