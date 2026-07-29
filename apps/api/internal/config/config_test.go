package config

import (
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/database"
)

func validConfig() Config {
	return Config{
		Port:            "8080",
		Environment:     "development",
		LogLevel:        "info",
		LogFormat:       "text",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 15 * time.Second,

		Database: database.Config{
			Host: "localhost",
			Port: 5432,

			User:     "postgres",
			Password: "postgres",
			Name:     "village",

			SSLMode: "disable",

			MaxConns:          20,
			MinConns:          2,
			MaxConnLifetime:   time.Hour,
			MaxConnIdleTime:   15 * time.Minute,
			HealthCheckPeriod: time.Minute,
		},
	}
}
