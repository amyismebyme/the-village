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
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
)

func Run() error {

	cfg := config.Load()
	//The service refuses to start with invalid configuration.
	if err := config.Validate(cfg); err != nil {
		return err
	}

	appLogger := logger.New(cfg)
	metrics.Register()
	httpServer := server.NewHTTPServer(appLogger, cfg)

	appLogger.Info("========================================")
	appLogger.Info(
		"Village API starting",
		"version", appruntime.BuildVersion,
		"go_version", appruntime.GoVersion(),
		"environment", cfg.Environment,
		"port", cfg.Port,
		"pid", os.Getpid(),
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
		appruntime.Uptime(),
	)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("shutdown signal received")
	appLogger.Info("Application started--", "uptime", appruntime.Uptime())
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
