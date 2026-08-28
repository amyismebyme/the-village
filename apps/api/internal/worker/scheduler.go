package worker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Scheduler struct {
	Interval time.Duration
}

func NewScheduler(
	interval time.Duration,
) (*Scheduler, error) {
	if interval <= 0 {
		return nil, errors.New(
			"worker scheduler: interval must be greater than zero",
		)
	}

	return &Scheduler{
		Interval: interval,
	}, nil
}

// Run executes fn immediately and returns the first function error.
func (s *Scheduler) Run(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return s.run(
		ctx,
		fn,
		func(err error) error {
			return err
		},
	)
}

// RunResilient executes fn immediately and continues scheduling even
// when an individual run fails.
//
// onError receives each run failure. It must not block indefinitely.
// Context cancellation remains a normal termination signal.
func (s *Scheduler) RunResilient(
	ctx context.Context,
	fn func(context.Context) error,
	onError func(error),
) error {
	if onError == nil {
		onError = func(error) {}
	}

	return s.run(
		ctx,
		fn,
		func(err error) error {
			onError(err)
			return nil
		},
	)
}

func (s *Scheduler) run(
	ctx context.Context,
	fn func(context.Context) error,
	handleError func(error) error,
) error {
	if ctx == nil {
		return errors.New(
			"worker scheduler: context is required",
		)
	}

	if fn == nil {
		return errors.New(
			"worker scheduler: function is required",
		)
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := fn(ctx); err != nil {
			if errors.Is(err, context.Canceled) &&
				ctx.Err() != nil {
				return ctx.Err()
			}

			if handledErr := handleError(err); handledErr != nil {
				return fmt.Errorf(
					"worker scheduler: run: %w",
					handledErr,
				)
			}
		}

		timer := time.NewTimer(s.Interval)

		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			return ctx.Err()

		case <-timer.C:
		}
	}
}
