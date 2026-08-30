package cache

import (
	"errors"
	"fmt"
	"strings"
)

type Key struct {
	Namespace string
	Source    string
	Operation string
	Identity  string
}

func (k Key) String() string {
	return strings.Join(
		[]string{
			k.Namespace,
			k.Source,
			k.Operation,
			k.Identity,
		},
		":",
	)
}

func (k Key) Validate() error {
	if strings.TrimSpace(k.Namespace) == "" {
		return errors.New(
			"cache key: namespace is required",
		)
	}

	if strings.TrimSpace(k.Source) == "" {
		return errors.New(
			"cache key: source is required",
		)
	}

	if strings.TrimSpace(k.Operation) == "" {
		return errors.New(
			"cache key: operation is required",
		)
	}

	if strings.TrimSpace(k.Identity) == "" {
		return errors.New(
			"cache key: identity is required",
		)
	}

	return nil
}

func CommunityKey(
	source string,
	communityID string,
) (string, error) {
	key := Key{
		Namespace: "community",
		Source:    source,
		Operation: "lookup",
		Identity:  communityID,
	}

	if err := key.Validate(); err != nil {
		return "", fmt.Errorf(
			"cache key: %w",
			err,
		)
	}

	return key.String(), nil
}

func RedditListingKey(
	subreddit string,
	after string,
	limit int,
) (string, error) {
	if strings.TrimSpace(subreddit) == "" {
		return "", errors.New(
			"cache key: subreddit is required",
		)
	}

	if limit <= 0 {
		return "", errors.New(
			"cache key: limit must be greater than zero",
		)
	}

	identity := fmt.Sprintf(
		"subreddit=%s,after=%s,limit=%d",
		subreddit,
		after,
		limit,
	)

	key := Key{
		Namespace: "reddit",
		Source:    "reddit",
		Operation: "listing",
		Identity:  identity,
	}

	return key.String(), nil
}
