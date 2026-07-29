package app

import (
	"log/slog"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
)

type Dependencies struct {
	Logger *slog.Logger



	DB     *database.Database
	Health *health.Registry
}
