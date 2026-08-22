package handlers

import (
	"log/slog"

	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

type Handler struct {
	communityService service.CommunityService
	logger           *slog.Logger
}

func NewHandler(
	communityService service.CommunityService,
	loggers ...*slog.Logger,
) *Handler {
	if communityService == nil {
		panic("handlers: community service is required")
	}

	var appLogger *slog.Logger

	if len(loggers) > 0 {
		appLogger = loggers[0]
	}

	return &Handler{
		communityService: communityService,
		logger:           appLogger,
	}
}
