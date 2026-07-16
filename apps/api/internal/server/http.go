package server
// Returns http server with timeouts defined in config.go
import (
	"net/http"
	"log/slog"
    "github.com/amyismebyme/the-village/apps/api/internal/config"
)

func NewHTTPServer(appLogger *slog.Logger,cfg config.Config) *http.Server {

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      NewRouter(appLogger),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
