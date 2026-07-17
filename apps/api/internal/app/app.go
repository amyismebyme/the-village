package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/logger"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
)

func Run() error {

	cfg := config.Load()
	appLogger := logger.New(cfg)

	httpServer := server.NewHTTPServer(appLogger, cfg)

	appLogger.Info("========================================")
	appLogger.Info(
		"Village API starting",
		"environment", cfg.Environment,
		"port", cfg.Port,
	)
	appLogger.Info("========================================")

	go func() {

		if err := httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	appLogger.Info(
		"server started successfully",
		"startup_ms",
		server.Uptime(),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("shutdown signal received")
	appLogger.Info("Total upTime: ", server.Uptime())

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)

	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	appLogger.Info("server shutdown complete")

	return nil
}
