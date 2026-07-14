package config

import "os"

// Config holds all application configuration.
type Config struct {
	Port        string
	Environment string
	LogLevel    string
}

func Load() Config {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	return cfg
}

// getEnv returns the value of an environment variable.
// If the variable is not set, it returns the supplied default.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}