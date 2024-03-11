package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/onrik/supermock/pkg/app"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
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

	server, err := app.New(*listen, *dbPath)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if err := server.Start(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
