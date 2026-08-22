package handlers

import (
	"net/http"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
)

func (h *Handler) logCommunityOperation(
	r *http.Request,
	operation string,
	communityID int64,
	status int,
	start time.Time,
) {
	if h.logger == nil {
		return
	}

	args := []any{
		"request_id",
		middleware.GetRequestID(r.Context()),
		"operation",
		operation,
		"status",
		status,
		"duration_ms",
		time.Since(start).Milliseconds(),
	}

	if communityID > 0 {
		args = append(
			args,
			"community_id",
			communityID,
		)
	}

	h.logger.Info(
		"community operation completed",
		args...,
	)
}
