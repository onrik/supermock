package main

import (
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	HttpAddr  string `env:"HTTP_ADDR" envDefault:"127.0.0.1:8000"`
	DB        string `env:"DB" envDefault:"sqlite://db.sqlite3"`
	SmtpAddr  string `env:"SMTP_ADDR"`
	SmtpDebug bool   `env:"SMTP_DEBUG"`
}

func (c *Config) slogLevel() slog.Level {
	var level slog.Level
	err := level.UnmarshalText([]byte(strings.ToUpper(c.LogLevel)))
	if err != nil {
		slog.Warn("Invalid log level", "level", c.LogLevel)
		level = slog.LevelInfo
	}

	return level
}

func readConfig() (Config, error) {
	config := Config{}
	err := env.Parse(&config)

	return config, err
}
