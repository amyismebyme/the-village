package database

import "context"

const CheckerName = "database"

type HealthProvider interface {
	Health(context.Context) error
}

type HealthChecker struct {
	db HealthProvider
}

func NewHealthChecker(db HealthProvider) *HealthChecker {
	return &HealthChecker{
		db: db,
	}
}

func (c *HealthChecker) Name() string {
	return CheckerName
}

func (c *HealthChecker) Check(ctx context.Context) error {
	return c.db.Health(ctx)
}
