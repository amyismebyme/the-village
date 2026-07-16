package logger

import (
	"log/slog"
	"os"
	"strings"
	"github.com/amyismebyme/the-village/apps/api/internal/config"
)

func New(cfg config.Config) *slog.Logger {

	level := slog.LevelInfo

	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	if strings.ToLower(cfg.LogFormat) == "json" {

		handler = slog.NewJSONHandler(os.Stdout, opts)

	} else {

		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}