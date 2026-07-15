package server
// Returns http server with timeouts defined in config.go
import (
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
)

func NewHTTPServer(cfg config.Config) *http.Server {

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      NewRouter(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
