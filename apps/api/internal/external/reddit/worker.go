package reddit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/worker"
)

const workerName = "reddit_ingestion"

type IngestionWorker struct {
	authenticator *Authenticator
	ingestion     *IngestionService
	scheduler     *worker.Scheduler

	subreddit string
	limit     int
	after     string

	logger *slog.Logger
}

type WorkerConfig struct {
	Subreddit string
	Limit     int
	After     string
	Interval  time.Duration
}

func NewIngestionWorker(
	authenticator *Authenticator,
	ingestion *IngestionService,
	config WorkerConfig,
) (*IngestionWorker, error) {
	if authenticator == nil {
		return nil, errors.New(
			"reddit worker: authenticator is required",
		)
	}

	if ingestion == nil {
		return nil, errors.New(
			"reddit worker: ingestion service is required",
		)
	}

	if config.Limit <= 0 {
		return nil, errors.New(
			"reddit worker: limit must be greater than zero",
		)
	}

	scheduler, err := worker.NewScheduler(
		config.Interval,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit worker: create scheduler: %w",
			err,
		)
	}

	return &IngestionWorker{
		authenticator: authenticator,
		ingestion:     ingestion,
		scheduler:     scheduler,
		subreddit:     config.Subreddit,
		limit:         config.Limit,
		after:         config.After,
	}, nil
}

func (w *IngestionWorker) SetLogger(
	logger *slog.Logger,
) {
	w.logger = logger
}

func (w *IngestionWorker) Run(
	ctx context.Context,
) error {
	return w.scheduler.RunResilient(
		ctx,
		w.runOnce,
		w.handleFailure,
	)
}

// RunOnce executes exactly one ingestion cycle.
//
// It is intentionally independent of the scheduler so integration tests
// and future administrative/manual execution can exercise one worker run
// deterministically.
func (w *IngestionWorker) RunOnce(
	ctx context.Context,
) error {
	return w.runOnce(ctx)
}

func (w *IngestionWorker) handleFailure(
	err error,
) {
	if w.logger == nil {
		return
	}

	w.logger.Error(
		"reddit worker run failed",
		"source",
		external.SourceReddit,
		"operation",
		"ingest",
		"error_type",
		redditStatusFromError(err),
	)
}

func (w *IngestionWorker) runOnce(
	ctx context.Context,
) error {
	start := time.Now()

	metrics.WorkersInFlight.
		WithLabelValues(workerName).
		Inc()

	defer metrics.WorkersInFlight.
		WithLabelValues(workerName).
		Dec()

	err := w.runOnceInternal(ctx)

	status := "success"

	if err != nil {
		// Cancellation caused by worker shutdown is normal termination,
		// not a failed worker execution.
		if errors.Is(err, context.Canceled) &&
			errors.Is(ctx.Err(), context.Canceled) {
			return err
		}

		status = "failure"

		metrics.WorkerFailuresTotal.
			WithLabelValues(workerName).
			Inc()
	}

	metrics.WorkerRunsTotal.
		WithLabelValues(
			workerName,
			status,
		).
		Inc()

	metrics.WorkerDuration.
		WithLabelValues(workerName).
		Observe(
			time.Since(start).Seconds(),
		)

	return err
}

func (w *IngestionWorker) runOnceInternal(
	ctx context.Context,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	token, err := w.authenticator.Token(ctx)
	if err != nil {
		return fmt.Errorf(
			"reddit worker: authenticate: %w",
			err,
		)
	}

	_, err = w.ingestion.IngestListing(
		ctx,
		token,
		w.subreddit,
		w.limit,
		w.after,
	)
	if err != nil {
		return fmt.Errorf(
			"reddit worker: ingest listing: %w",
			err,
		)
	}

	return nil
}
