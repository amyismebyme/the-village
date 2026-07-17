package server

// Returns http server with timeouts defined in config.go
import (
	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"log/slog"
	"net/http"
)

func NewHTTPServer(appLogger *slog.Logger, cfg config.Config) *http.Server {

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      NewRouter(appLogger),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
