package reddit

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func observeOperation(
	logger *slog.Logger,
	operation string,
	externalID string,
	status string,
	start time.Time,
	err error,
) {
	external.ObserveOperation(
		logger,
		external.SourceReddit,
		operation,
		externalID,
		status,
		start,
		err,
	)
}

func redditStatusFromError(err error) string {
	switch {
	case errors.Is(err, external.ErrUnauthorized):
		return "401"

	case errors.Is(err, external.ErrForbidden):
		return "403"

	case errors.Is(err, external.ErrNotFound):
		return "404"

	case errors.Is(err, external.ErrRateLimited):
		return "429"

	case errors.Is(err, external.ErrTimeout):
		return "timeout"

	case errors.Is(err, external.ErrUpstream):
		return "5xx"

	case errors.Is(err, external.ErrInvalidPayload):
		return "invalid_payload"

	case errors.Is(err, external.ErrInvalidConfig):
		return "invalid_config"

	case errors.Is(err, context.Canceled):
		return "canceled"

	default:
		return "error"
	}
}
