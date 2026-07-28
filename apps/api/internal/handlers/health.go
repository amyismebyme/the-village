package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/health"
)

var healthRegistry *health.Registry

// SetHealthRegistry injects the application's health registry.
func SetHealthRegistry(registry *health.Registry) {
	healthRegistry = registry
}

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	response := HealthResponse{
		Status: "healthy",
		Checks: make(map[string]string),
	}

	// Registry not configured.
	if healthRegistry == nil {

		response.Status = "unhealthy"
		response.Checks["registry"] = "not configured"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("failed to encode health response: %v", err)
		}

		return
	}

	results := healthRegistry.Check(r.Context())

	for name, err := range results {

		if err != nil {

			response.Status = "unhealthy"
			response.Checks[name] = err.Error()

			continue
		}

		response.Checks[name] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")

	if response.Status == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {

		log.Printf("failed to encode health response: %v", err)

		http.Error(
			w,
			"failed to encode health response",
			http.StatusInternalServerError,
		)

		return
	}
}
