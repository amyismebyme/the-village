package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/logger"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

func Run() error {

	cfg := config.Load()

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}

	if err := cfg.Database.Validate(); err != nil {
		return fmt.Errorf("database configuration: %w", err)
	}

	appLogger := logger.New(cfg)

	startupCtx := context.Background()

	db, err := database.Open(startupCtx, cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	startupComplete := false

	defer func() {
		if !startupComplete {
			db.Close()
		}
	}()

	//------------------------------------------------------------------
	// Health
	//------------------------------------------------------------------

	healthRegistry := health.NewRegistry()

	healthRegistry.Register(
		database.NewHealthChecker(db),
	)

	//------------------------------------------------------------------
	// Metrics
	//------------------------------------------------------------------

	stats := db.Stats()

	appLogger.Info(
		"database connection pool initialized",
		"max_connections", stats.MaxConnections,
		"total_connections", stats.TotalConnections,
		"idle_connections", stats.IdleConnections,
	)

	metrics.Register(
		prometheus.DefaultRegisterer,
		db.Pool(),
	)

	//------------------------------------------------------------------
	// Dependency Injection
	//------------------------------------------------------------------

	communityRepository := postgres.NewCommunityRepository(
		db.Pool(),
	)

	communityService := service.NewCommunityService(
		communityRepository,
	)

	handler := handlers.NewHandler(
		communityService,
	)

	//------------------------------------------------------------------
	// HTTP Server
	//------------------------------------------------------------------

	httpServer := server.NewHTTPServer(
		appLogger,
		cfg,
		healthRegistry,
		handler,
	)

	appLogger.Info(
		"Village API starting",
		"version", appruntime.BuildVersion,
		"go_version", appruntime.GoVersion(),
		"environment", cfg.Environment,
		"port", cfg.Port,
		"pid", os.Getpid(),
	)

	serverErrors := make(chan error, 1)

	go func() {

		err := httpServer.ListenAndServe()

		if err != nil &&
			err != http.ErrServerClosed {

			serverErrors <- err
			return
		}

		serverErrors <- nil
	}()

	appLogger.Info(
		"server started successfully",
		"startup_ms",
		appruntime.Uptime().Milliseconds(),
	)

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer signal.Stop(stop)

	startupComplete = true

	select {

	case sig := <-stop:

		appLogger.Info(
			"shutdown signal received",
			"signal", sig.String(),
			"uptime", appruntime.Uptime().String(),
		)

	case err := <-serverErrors:

		if err != nil {
			db.Close()
			return fmt.Errorf(
				"http server failed: %w",
				err,
			)
		}

		db.Close()
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)

	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {

		db.Close()

		return fmt.Errorf(
			"shutdown http server: %w",
			err,
		)
	}

	db.Close()

	appLogger.Info(
		"server shutdown complete",
	)

	return nil
}
