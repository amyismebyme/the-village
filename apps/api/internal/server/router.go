package server

import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
)

func NewRouter() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/ready", handlers.ReadyHandler)
	mux.HandleFunc("/version", handlers.VersionHandler)

	return mux
}
