package ratelimit

import (
	"errors"
	"sync"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

type PerSourceWaitObserver func(
	source external.Source,
	waited time.Duration,
	err error,
)

type PerSource struct {
	mu       sync.Mutex
	limiters map[external.Source]external.RateLimiter
	observer PerSourceWaitObserver
}

func NewPerSource() *PerSource {
	return &PerSource{
		limiters: make(
			map[external.Source]external.RateLimiter,
		),
	}
}

// SetObserver sets an observer that receives wait events from all
// registered source-specific pacers.
//
// The source is included so callers can determine which external
// source produced the event.
func (p *PerSource) SetObserver(observer PerSourceWaitObserver) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.observer = observer

	for source, limiter := range p.limiters {
		p.attachObserver(source, limiter)
	}
}

func (p *PerSource) attachObserver(
	source external.Source,
	limiter external.RateLimiter,
) {
	pacer, ok := limiter.(*Pacer)
	if !ok {
		return
	}

	if p.observer == nil {
		pacer.SetObserver(nil)
		return
	}

	observer := p.observer

	pacer.SetObserver(func(
		waited time.Duration,
		err error,
	) {
		observer(source, waited, err)
	})
}

func (p *PerSource) Register(
	source external.Source,
	interval time.Duration,
) (external.RateLimiter, error) {
	if source == "" {
		return nil, errors.New(
			"rate limiter: source is required",
		)
	}

	limiter, err := NewPacer(interval)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.limiters[source]; exists {
		return nil, errors.New(
			"rate limiter: source already registered",
		)
	}

	p.limiters[source] = limiter
	p.attachObserver(source, limiter)

	return limiter, nil
}

func (p *PerSource) Get(
	source external.Source,
) (external.RateLimiter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	limiter, ok := p.limiters[source]

	return limiter, ok
}
