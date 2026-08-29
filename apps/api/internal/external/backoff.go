package external

import (
	"errors"
	"math/rand"
	"time"
)

type Backoff struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	Jitter     float64
}

func NewBackoff(
	initial time.Duration,
	max time.Duration,
	multiplier float64,
	jitter float64,
) (Backoff, error) {
	if initial <= 0 {
		return Backoff{}, errors.New(
			"backoff: initial delay must be greater than zero",
		)
	}

	if max < initial {
		return Backoff{}, errors.New(
			"backoff: max delay must be greater than or equal to initial delay",
		)
	}

	if multiplier < 1 {
		return Backoff{}, errors.New(
			"backoff: multiplier must be at least 1",
		)
	}

	if jitter < 0 || jitter > 1 {
		return Backoff{}, errors.New(
			"backoff: jitter must be between 0 and 1",
		)
	}

	return Backoff{
		Initial:    initial,
		Max:        max,
		Multiplier: multiplier,
		Jitter:     jitter,
	}, nil
}

func (b Backoff) Delay(
	attempt int,
) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	delay := float64(b.Initial)

	for i := 1; i < attempt; i++ {
		delay *= b.Multiplier

		if delay >= float64(b.Max) {
			delay = float64(b.Max)
			break
		}
	}

	result := time.Duration(delay)

	if result > b.Max {
		result = b.Max
	}

	if b.Jitter == 0 {
		return result
	}

	factor := 1 +
		((rand.Float64()*2 - 1) * b.Jitter)

	result = time.Duration(
		float64(result) * factor,
	)

	if result < 0 {
		return 0
	}

	if result > b.Max {
		return b.Max
	}

	return result
}
