package config

// Package config manages application configuration.
// Manages different params like ports and timeouts and log settings
import (
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
