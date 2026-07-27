package database

import (
	"fmt"
	"time"
)

type Config struct {
	Host string
	Port int

	User     string
	Password string
	Name     string

	SSLMode string

	MaxConns int32
	MinConns int32

	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration

	HealthCheckPeriod time.Duration
}

func (c Config) DSN() string {

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

func (c Config) Validate() error {

	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Port == 0 {
		return fmt.Errorf("database port is required")
	}

	if c.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Name == "" {
		return fmt.Errorf("database name is required")
	}

	return nil
}
