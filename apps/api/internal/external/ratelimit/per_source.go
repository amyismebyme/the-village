package ratelimit

import (
	"errors"
	"sync"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

type PerSource struct {
	mu       sync.Mutex
	limiters map[external.Source]external.RateLimiter
}

func NewPerSource() *PerSource {
	return &PerSource{
		limiters: make(
			map[external.Source]external.RateLimiter,
		),
	}
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
