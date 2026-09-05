package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

type WaitObserver func(waited time.Duration, err error)

type Pacer struct {
	interval time.Duration

	mu       sync.Mutex
	last     time.Time
	observer WaitObserver
}

func NewPacer(
	interval time.Duration,
) (*Pacer, error) {
	if interval <= 0 {
		return nil, errors.New(
			"rate limiter: interval must be greater than zero",
		)
	}

	return &Pacer{
		interval: interval,
	}, nil
}

// SetObserver sets the callback invoked after each Wait attempt.
//
// The observer is called after the pacer's mutex has been released,
// so observers may safely interact with the pacer or other state.
func (p *Pacer) SetObserver(observer WaitObserver) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.observer = observer
}

// observe notifies the current observer, if one is configured.
//
// The observer is copied while holding the mutex and invoked after
// releasing it to avoid holding the pacer's lock during callbacks.
func (p *Pacer) observe(
	waited time.Duration,
	err error,
) {
	p.mu.Lock()
	observer := p.observer
	p.mu.Unlock()

	if observer != nil {
		observer(waited, err)
	}
}

// Wait blocks until another request may start.
//
// The first request proceeds immediately. Every later request is
// separated by at least interval. Cancellation while waiting does
// not consume a request slot.
func (p *Pacer) Wait(
	ctx context.Context,
) error {
	if ctx == nil {
		err := errors.New(
			"rate limiter: context is required",
		)

		p.observe(0, err)

		return err
	}

	start := time.Now()

	p.mu.Lock()

	if err := ctx.Err(); err != nil {
		p.mu.Unlock()

		p.observe(time.Since(start), err)

		return err
	}

	if !p.last.IsZero() {
		wait := p.interval - time.Since(p.last)

		if wait > 0 {
			timer := time.NewTimer(wait)

			select {
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}

				p.mu.Unlock()

				err := ctx.Err()
				p.observe(time.Since(start), err)

				return err

			case <-timer.C:
			}
		}
	}

	if err := ctx.Err(); err != nil {
		p.mu.Unlock()

		p.observe(time.Since(start), err)

		return err
	}

	p.last = time.Now()

	p.mu.Unlock()

	p.observe(time.Since(start), nil)

	return nil
}
