package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Lifecycle struct {
	worker Worker

	mu sync.Mutex

	running bool
	cancel  context.CancelFunc
	done    chan struct{}
	runErr  error
}

func NewLifecycle(worker Worker) *Lifecycle {
	if worker == nil {
		panic(
			"worker lifecycle: worker is required",
		)
	}

	return &Lifecycle{
		worker: worker,
	}
}

// Run starts the worker and blocks until the worker exits.
//
// The supplied context remains authoritative for cancellation and
// deadlines. Lifecycle adds a child cancellation context so Stop can
// request graceful termination.
func (l *Lifecycle) Run(
	ctx context.Context,
) error {
	if ctx == nil {
		return errors.New(
			"worker lifecycle: context is required",
		)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	l.mu.Lock()

	if l.running {
		l.mu.Unlock()

		return errors.New(
			"worker lifecycle: worker already running",
		)
	}

	runCtx, cancel := context.WithCancel(ctx)

	l.running = true
	l.cancel = cancel
	l.done = make(chan struct{})
	l.runErr = nil

	done := l.done

	l.mu.Unlock()

	err := l.worker.Run(runCtx)

	cancel()

	l.mu.Lock()

	l.running = false
	l.cancel = nil
	l.runErr = err

	close(done)

	l.mu.Unlock()

	if err != nil {
		return fmt.Errorf(
			"worker lifecycle: run worker: %w",
			err,
		)
	}

	return nil
}

// Stop requests worker cancellation and waits for it to exit.
//
// The supplied context bounds how long the caller is willing to wait
// for graceful termination.
func (l *Lifecycle) Stop(
	ctx context.Context,
) error {
	if ctx == nil {
		return errors.New(
			"worker lifecycle: stop context is required",
		)
	}

	l.mu.Lock()

	if !l.running {
		l.mu.Unlock()

		return nil
	}

	cancel := l.cancel
	done := l.done

	l.mu.Unlock()

	cancel()

	select {
	case <-done:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Lifecycle) Running() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.running
}
