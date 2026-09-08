package reddit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const workerName = "reddit_ingestion"

type IngestionWorker struct {
	authenticator  *Authenticator
	ingestion      *IngestionService
	itemRepository repository.ExternalItemRepository
	scheduler      *worker.Scheduler

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
	itemRepository repository.ExternalItemRepository,
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

	if itemRepository == nil {
		return nil, errors.New(
			"reddit worker: item repository is required",
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
		authenticator:  authenticator,
		ingestion:      ingestion,
		itemRepository: itemRepository,
		scheduler:      scheduler,
		subreddit:      config.Subreddit,
		limit:          config.Limit,
		after:          config.After,
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
		string(external.ClassifyError(err)),
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
		// Worker shutdown cancellation is normal termination and should
		// not be counted as a worker failure.
		if errors.Is(err, context.Canceled) &&
			errors.Is(ctx.Err(), context.Canceled) {
			return err
		}

		status = "failure"

		metrics.WorkerFailuresTotal.
			WithLabelValues(workerName).
			Inc()

		metrics.WorkerFailureTypesTotal.
			WithLabelValues(
				workerName,
				string(external.ClassifyError(err)),
			).
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
) (err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "reddit.ingest")
	span.SetAttributes(
		attribute.String("source", string(external.SourceReddit)),
		attribute.String("operation", "ingest"),
		attribute.String("subreddit", w.subreddit),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "reddit ingestion failed")
		}
		span.End()
	}()

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

	items, err := w.ingestion.IngestListing(
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

	if err := ctx.Err(); err != nil {
		return err
	}

	persistCtx, persistSpan := otel.Tracer(tracerName).Start(
		ctx,
		"external_item.upsert_batch",
	)
	persistSpan.SetAttributes(
		attribute.String("source", string(external.SourceReddit)),
		attribute.String("operation", "upsert_batch"),
	)

	if err := w.itemRepository.UpsertBatch(
		persistCtx,
		items,
	); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, "external item persistence failed")
		persistSpan.End()

		return fmt.Errorf(
			"reddit worker: persist ingested items: %w",
			err,
		)
	}

	persistSpan.End()

	return nil
}
