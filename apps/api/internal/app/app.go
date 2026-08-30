package app

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/amyismebyme/the-village/apps/api/internal/cache"
	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/external/ratelimit"
	"github.com/amyismebyme/the-village/apps/api/internal/external/reddit"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/logger"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	appruntime "github.com/amyismebyme/the-village/apps/api/internal/runtime"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
	"github.com/amyismebyme/the-village/apps/api/internal/worker"
)

func Run() error {
	cfg := config.Load()

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf(
			"validate configuration: %w",
			err,
		)
	}

	if err := cfg.Database.Validate(); err != nil {
		return fmt.Errorf(
			"database configuration: %w",
			err,
		)
	}

	appLogger := logger.New(cfg)

	startupCtx := context.Background()

	db, err := database.Open(
		startupCtx,
		cfg.Database,
	)
	if err != nil {
		return fmt.Errorf(
			"open database: %w",
			err,
		)
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
		"max_connections",
		stats.MaxConnections,
		"total_connections",
		stats.TotalConnections,
		"idle_connections",
		stats.IdleConnections,
	)

	if err := metrics.Register(
		prometheus.DefaultRegisterer,
		db.Pool(),
	); err != nil {
		return fmt.Errorf(
			"register metrics: %w",
			err,
		)
	}

	//------------------------------------------------------------------
	// Cache
	//------------------------------------------------------------------

	var (
		applicationCache cache.Cache
		memoryCache      *cache.Memory
	)

	if cfg.Cache.Enabled {
		var err error

		memoryCache, err = cache.NewMemory(
			cfg.Cache.MaxEntries,
		)
		if err != nil {
			return fmt.Errorf(
				"create application cache: %w",
				err,
			)
		}

		applicationCache = memoryCache

		if err := metrics.RegisterCache(
			prometheus.DefaultRegisterer,
			func() metrics.CacheStats {
				stats := memoryCache.Stats()

				return metrics.CacheStats{
					Entries:   stats.Entries,
					Hits:      stats.Hits,
					Misses:    stats.Misses,
					Evictions: stats.Evictions,
				}
			},
		); err != nil {
			return fmt.Errorf(
				"register cache metrics: %w",
				err,
			)
		}

		appLogger.Info(
			"application cache initialized",
			"max_entries",
			cfg.Cache.MaxEntries,
			"reddit_listing_ttl_seconds",
			int64(
				cfg.Cache.RedditListingTTL.Seconds(),
			),
		)
	}

	//------------------------------------------------------------------
	// Dependency Injection
	//------------------------------------------------------------------

	communityRepository := postgres.NewCommunityRepository(
		db.Pool(),
		cfg.Database.QueryTimeout,
	)

	communityService := service.NewCommunityService(
		communityRepository,
	)

	handler := handlers.NewHandler(
		communityService,
		appLogger,
	)

	//------------------------------------------------------------------
	// Background Workers
	//------------------------------------------------------------------

	workerRuntime := worker.NewRuntime()

	if cfg.Worker.Enabled {
		if !cfg.External.Reddit.Enabled {
			return fmt.Errorf(
				"worker configuration: Reddit must be enabled when workers are enabled",
			)
		}

		// -------------------------------------------------------------
		// External retry policy
		// -------------------------------------------------------------

		retryBackoff, err := external.NewBackoff(
			cfg.External.Retry.InitialBackoff,
			cfg.External.Retry.MaxBackoff,
			cfg.External.Retry.BackoffMultiplier,
			cfg.External.Retry.Jitter,
		)
		if err != nil {
			return fmt.Errorf(
				"configure external retry backoff: %w",
				err,
			)
		}

		retryPolicy, err := external.NewRetryPolicy(
			cfg.External.Retry.MaxAttempts,
			retryBackoff,
		)
		if err != nil {
			return fmt.Errorf(
				"configure external retry policy: %w",
				err,
			)
		}

		// -------------------------------------------------------------
		// Per-source rate limiter
		// -------------------------------------------------------------

		rateLimiters := ratelimit.NewPerSource()

		redditLimiter, err := rateLimiters.Register(
			external.SourceReddit,
			cfg.External.Reddit.RequestInterval,
		)
		if err != nil {
			return fmt.Errorf(
				"configure Reddit rate limiter: %w",
				err,
			)
		}

		// -------------------------------------------------------------
		// Reddit HTTP client
		// -------------------------------------------------------------

		redditHTTPClient := &http.Client{}

		redditAuthenticator, err := reddit.NewAuthenticator(
			redditHTTPClient,
			cfg.External.Reddit.AuthBaseURL,
			cfg.External.Reddit.ClientID,
			cfg.External.Reddit.ClientSecret,
			cfg.External.Reddit.UserAgent,
			cfg.External.RequestTimeout,
		)
		if err != nil {
			return fmt.Errorf(
				"create Reddit authenticator: %w",
				err,
			)
		}

		redditClient, err := reddit.NewClient(
			redditHTTPClient,
			cfg.External.Reddit.BaseURL,
			cfg.External.Reddit.UserAgent,
			cfg.External.RequestTimeout,
		)
		if err != nil {
			return fmt.Errorf(
				"create Reddit client: %w",
				err,
			)
		}

		// -------------------------------------------------------------
		// Reddit observability + rate limiting + retry
		// -------------------------------------------------------------

		redditAuthenticator.SetLogger(
			appLogger,
		)

		redditAuthenticator.SetRateLimiter(
			redditLimiter,
		)

		redditAuthenticator.SetRetryPolicy(
			retryPolicy,
		)

		redditClient.SetLogger(
			appLogger,
		)

		redditClient.SetRateLimiter(
			redditLimiter,
		)

		redditClient.SetRetryPolicy(
			retryPolicy,
		)

		// -------------------------------------------------------------
		// Reddit ingestion
		// -------------------------------------------------------------

		redditIngestion := reddit.NewIngestionService(
			redditClient,
			reddit.NewPostNormalizer(),
		)

		if applicationCache != nil {
			if err := redditIngestion.SetCache(
				applicationCache,
				cfg.Cache.RedditListingTTL,
			); err != nil {
				return fmt.Errorf(
					"configure Reddit cache: %w",
					err,
				)
			}
		}

		redditWorker, err := reddit.NewIngestionWorker(
			redditAuthenticator,
			redditIngestion,
			reddit.WorkerConfig{
				Subreddit: cfg.Worker.Reddit.Subreddit,
				Limit:     cfg.Worker.Reddit.Limit,
				After:     "",
				Interval:  cfg.Worker.Reddit.IngestInterval,
			},
		)
		if err != nil {
			return fmt.Errorf(
				"create Reddit worker: %w",
				err,
			)
		}

		redditWorker.SetLogger(
			appLogger,
		)

		redditLifecycle := worker.NewLifecycle(
			redditWorker,
		)

		if err := workerRuntime.Add(
			redditLifecycle,
		); err != nil {
			return fmt.Errorf(
				"register Reddit worker: %w",
				err,
			)
		}
	}

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
		"version",
		appruntime.BuildVersion,
		"go_version",
		appruntime.GoVersion(),
		"environment",
		cfg.Environment,
		"port",
		cfg.Port,
		"pid",
		os.Getpid(),
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

	workerCtx, workerCancel := context.WithCancel(
		context.Background(),
	)
	defer workerCancel()

	if cfg.Worker.Enabled {
		if err := workerRuntime.Start(
			workerCtx,
		); err != nil {
			db.Close()

			return fmt.Errorf(
				"start workers: %w",
				err,
			)
		}

		appLogger.Info(
			"background workers started",
		)
	}

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
			"signal",
			sig.String(),
			"uptime",
			appruntime.Uptime().String(),
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

	//------------------------------------------------------------------
	// Graceful shutdown
	//------------------------------------------------------------------

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	// Stop workers before the HTTP server because workers may depend
	// on application resources during their final operation.
	if err := workerRuntime.Stop(
		shutdownCtx,
	); err != nil {
		db.Close()

		return fmt.Errorf(
			"shutdown workers: %w",
			err,
		)
	}

	if err := httpServer.Shutdown(
		shutdownCtx,
	); err != nil {
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
