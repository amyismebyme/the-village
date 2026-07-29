package config

// Package config manages application configuration.
// Manages different params like ports and timeouts and log settings
import (
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Port            string
	Environment     string
	LogLevel        string
	LogFormat       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	Database        database.Config
}

// Load reads environment variables.
func Load() Config {

	cfg := Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "text"),
		ReadTimeout:     getDuration("READ_TIMEOUT", 10),
		WriteTimeout:    getDuration("WRITE_TIMEOUT", 10),
		IdleTimeout:     getDuration("IDLE_TIMEOUT", 60),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 15),

		Database: database.Config{
			Host: getEnv("DB_HOST", "postgres"),
			Port: getInt("DB_PORT", 5432),

			User:     getEnv("DB_USER", "village"),
			Password: getEnv("DB_PASSWORD", "village"),
			Name:     getEnv("DB_NAME", "village"),

			SSLMode: getEnv("DB_SSLMODE", "disable"),

			MaxConns:          int32(getInt("DB_MAX_CONNS", 10)),
			MinConns:          int32(getInt("DB_MIN_CONNS", 1)),
			MaxConnLifetime:   getDuration("DB_MAX_CONN_LIFETIME", 3600),
			MaxConnIdleTime:   getDuration("DB_MAX_CONN_IDLE_TIME", 300),
			HealthCheckPeriod: getDuration("DB_HEALTH_CHECK_PERIOD", 60),
		},
	}

	return cfg
}

func getEnv(key, defaultValue string) string {

	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func getDuration(key string, defaultSeconds int) time.Duration {

	value := os.Getenv(key)

	if value == "" {
		return time.Duration(defaultSeconds) * time.Second
	}

	seconds, err := strconv.Atoi(value)

	if err != nil {
		return time.Duration(defaultSeconds) * time.Second
	}

	return time.Duration(seconds) * time.Second
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return v
}
