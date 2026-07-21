package testutil

import (
	"net/http"
	"log/slog"

	"github.com/amyismebyme/the-village/apps/api/internal/server"
)

func NewRouter(appLogger *slog.Logger) http.Handler {
    return server.NewRouter(appLogger)
}