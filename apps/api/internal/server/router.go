package server
// Route the different events based on handlers
import (
	"net/http"
    "log/slog"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
)

func NewRouter(appLogger *slog.Logger) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/ready", handlers.ReadyHandler)
	mux.HandleFunc("/version", handlers.VersionHandler)

	handler := middleware.Recovery(
    	appLogger,
    	middleware.RequestID(
    		middleware.Logging(appLogger, mux),
    	),
    )

	return handler
}