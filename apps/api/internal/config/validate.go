package config

import (
	"fmt"
	"strings"
)

// Validates params at application startup time and exits if not matching
func Validate(cfg Config) error {

	if cfg.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}

	switch strings.ToLower(cfg.Environment) {
	case "development", "staging", "production":
	default:
		return fmt.Errorf("invalid environment: %s", cfg.Environment)
	}

	switch strings.ToLower(cfg.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}

	switch strings.ToLower(cfg.LogFormat) {
	case "text", "json":
	default:
		return fmt.Errorf("invalid log format: %s", cfg.LogFormat)
	}

	return nil
}
