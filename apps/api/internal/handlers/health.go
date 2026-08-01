package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/health"
)

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// NewHealthHandler returns a health endpoint backed by the supplied registry.
func NewHealthHandler(
	appLogger *slog.Logger,
	registry *health.Registry,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status: "healthy",
			Checks: make(map[string]string),
		}

		statusCode := http.StatusOK

		if registry == nil {
			response.Status = "unhealthy"
			response.Checks["registry"] = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		} else {
			for name, checkErr := range registry.Check(r.Context()) {
				if checkErr != nil {
					response.Status = "unhealthy"
					response.Checks[name] = "unhealthy"
					statusCode = http.StatusServiceUnavailable

					if appLogger != nil {
						appLogger.Warn(
							"health check failed",
							"check", name,
							"error", checkErr,
						)
					}

					continue
				}

				response.Checks[name] = "healthy"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			if appLogger != nil {
				appLogger.Error(
					"failed to encode health response",
					"error", err,
				)
			}
		}
	})
}
