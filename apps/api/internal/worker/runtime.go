package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Runtime struct {
	mu sync.Mutex

	lifecycles []*Lifecycle

	wg sync.WaitGroup

	errCh chan error
}

func NewRuntime() *Runtime {
	return &Runtime{
		errCh: make(chan error, 16),
	}
}

func (r *Runtime) Add(
	lifecycle *Lifecycle,
) error {
	if lifecycle == nil {
		return errors.New(
			"worker runtime: lifecycle is required",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.lifecycles = append(
		r.lifecycles,
		lifecycle,
	)

	return nil
}

func (r *Runtime) Start(
	ctx context.Context,
) error {
	if ctx == nil {
		return errors.New(
			"worker runtime: context is required",
		)
	}

	r.mu.Lock()

	lifecycles := append(
		[]*Lifecycle(nil),
		r.lifecycles...,
	)

	r.mu.Unlock()

	for _, lifecycle := range lifecycles {
		r.wg.Add(1)

		go func(lc *Lifecycle) {
			defer r.wg.Done()

			err := lc.Run(ctx)

			if err == nil ||
				errors.Is(
					err,
					context.Canceled,
				) {
				return
			}

			select {
			case r.errCh <- err:
			default:
				// Do not allow one broken worker to block the
				// runtime forever trying to report an error.
			}
		}(lifecycle)
	}

	return nil
}

func (r *Runtime) Stop(
	ctx context.Context,
) error {
	if ctx == nil {
		return errors.New(
			"worker runtime: stop context is required",
		)
	}

	r.mu.Lock()

	lifecycles := append(
		[]*Lifecycle(nil),
		r.lifecycles...,
	)

	r.mu.Unlock()

	var firstErr error

	for _, lifecycle := range lifecycles {
		if err := lifecycle.Stop(ctx); err != nil &&
			firstErr == nil {
			firstErr = fmt.Errorf(
				"stop worker: %w",
				err,
			)
		}
	}

	r.wg.Wait()

	return firstErr
}

func (r *Runtime) Errors() <-chan error {
	return r.errCh
}
