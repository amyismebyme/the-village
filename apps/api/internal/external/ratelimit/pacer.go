package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Pacer struct {
	interval time.Duration

	mu   sync.Mutex
	last time.Time
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

// Wait blocks until another request may start.
//
// The first request proceeds immediately. Every later request is
// separated by at least interval. Cancellation while waiting does
// not consume a request slot.
func (p *Pacer) Wait(
	ctx context.Context,
) error {
	if ctx == nil {
		return errors.New(
			"rate limiter: context is required",
		)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
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

				return ctx.Err()

			case <-timer.C:
			}
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	p.last = time.Now()

	return nil
}
