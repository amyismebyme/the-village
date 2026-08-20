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
	QueryTimeout     time.Duration
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

	if c.MaxConns <= 0 {
		return fmt.Errorf("database max connections must be greater than zero")
	}

	if c.MinConns < 0 {
		return fmt.Errorf("database min connections cannot be negative")
	}

	if c.MinConns > c.MaxConns {
		return fmt.Errorf("database min connections cannot exceed max connections")
	}

	if c.MaxConnLifetime <= 0 {
		return fmt.Errorf("database max connection lifetime must be greater than zero")
	}

	if c.MaxConnIdleTime <= 0 {
		return fmt.Errorf("database max idle time must be greater than zero")
	}

	if c.HealthCheckPeriod <= 0 {
		return fmt.Errorf("database health check period must be greater than zero")
	}

    if c.QueryTimeout <= 0 {
    	return fmt.Errorf(
    		"database query timeout must be greater than zero",
    	)
    }

	return nil
}
