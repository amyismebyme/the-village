package health

import (
	"log/slog"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
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

func (h *HealthHandler) handleLiveness(
	w http.ResponseWriter,
	r *http.Request,
) {
	httputil.WriteJSON(
		w,
		http.StatusOK,
		HealthResponse{
			Status: "healthy",
		},
	)
}

func (h *HealthHandler) handleReadiness(
	w http.ResponseWriter,
	r *http.Request,
) {
	if h.registry == nil {
		httputil.WriteJSON(
			w,
			http.StatusServiceUnavailable,
			HealthResponse{
				Status: "unhealthy",
			},
		)

		return
	}

	results := h.registry.Check(r.Context())

	healthy := len(results) > 0

	for _, result := range results {
		if result.Error != "" {
			healthy = false
			break
		}
	}

	if !healthy {
		httputil.WriteJSON(
			w,
			http.StatusServiceUnavailable,
			HealthResponse{
				Status: "unhealthy",
				Checks: results,
			},
		)

		return
	}

	httputil.WriteJSON(
		w,
		http.StatusOK,
		HealthResponse{
			Status: "ready",
			Checks: results,
		},
	)
}
