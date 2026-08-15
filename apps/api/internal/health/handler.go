package health

import (
	"log/slog"
	"net/http"
)

type HealthResponse struct {
	Status string   `json:"status"`
	Checks []Result `json:"checks,omitempty"`
}

type HealthHandler struct {
	logger   *slog.Logger
	registry *Registry
}

func NewHealthHandler(
	logger *slog.Logger,
	registry *Registry,
) http.Handler {
	return &HealthHandler{
		logger:   logger,
		registry: registry,
	}
}

// ServeHTTP exposes the application's liveness and readiness endpoints.
//
// /health
//
//	Liveness check.
//	Does not inspect external dependencies such as PostgreSQL.
//
// /ready
//
//	Readiness check.
//	Verifies registered dependencies before allowing traffic.
func (h *HealthHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.URL.Path {
	case "/health":
		h.handleLiveness(w, r)

	case "/ready":
		h.handleReadiness(w, r)

	default:
		http.NotFound(w, r)
	}
}

// handleLiveness answers:
//
//	"Is the API process alive and able to serve requests?"
//
// PostgreSQL and other external dependencies must NOT be checked here.
//
// A temporary database outage should therefore not cause Kubernetes/Docker
// to consider the API process dead and unnecessarily restart the container.
func (h *HealthHandler) handleLiveness(
	w http.ResponseWriter,
	_ *http.Request,
) {
	writeJSON(
		w,
		http.StatusOK,
		HealthResponse{
			Status: "healthy",
		},
	)
}

// handleReadiness answers:
//
//	"Can this API instance safely receive traffic?"
//
// Unlike liveness, readiness checks registered dependencies such as
// PostgreSQL.
func (h *HealthHandler) handleReadiness(
	w http.ResponseWriter,
	r *http.Request,
) {
	// No registry means that the API cannot establish dependency
	// readiness.
	if h.registry == nil {
		writeJSON(
			w,
			http.StatusServiceUnavailable,
			HealthResponse{
				Status: "unhealthy",
			},
		)

		return
	}

	results := h.registry.Check(r.Context())

	healthy := true

	for _, result := range results {
		if result.Error != "" {
			healthy = false
			break
		}
	}

	if !healthy {
		writeJSON(
			w,
			http.StatusServiceUnavailable,
			HealthResponse{
				Status: "unhealthy",
				Checks: results,
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		HealthResponse{
			Status: "ready",
			Checks: results,
		},
	)
}
