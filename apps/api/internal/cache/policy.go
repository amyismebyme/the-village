package cache

import (
	"errors"
	"time"
)

type TTLPolicy struct {
	RedditListing time.Duration
	Community     time.Duration
}

func DefaultTTLPolicy() TTLPolicy {
	return TTLPolicy{
		RedditListing: 2 * time.Minute,
		Community:     5 * time.Minute,
	}
}

func (p TTLPolicy) Validate() error {
	if p.RedditListing <= 0 {
		return errors.New(
			"cache policy: Reddit listing TTL must be greater than zero",
		)
	}

	if p.Community <= 0 {
		return errors.New(
			"cache policy: Community TTL must be greater than zero",
		)
	}

	return nil
}
