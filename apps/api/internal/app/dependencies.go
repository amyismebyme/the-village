package app

import (
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"log/slog"
)

type Dependencies struct {
	Logger *slog.Logger

	DB     *database.Database
	Health *health.Registry
}
