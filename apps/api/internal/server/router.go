package server

import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
)

func NewRouter() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/ready", handlers.ReadyHandler)
	mux.HandleFunc("/version", handlers.VersionHandler)

	handler := middleware.Recovery(
		middleware.RequestID(
			middleware.Logging(mux),
		),
	)

	return handler
}