package reddit

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"net/url"
	"strings"
)

type PostNormalizer struct{}

var _ external.Normalizer[RedditPost] = PostNormalizer{}

func NewPostNormalizer() PostNormalizer {
	return PostNormalizer{}
}

func (PostNormalizer) Normalize(
	ctx context.Context,
	input RedditPost,
) (external.Item, error) {
	if err := ctx.Err(); err != nil {
		return external.Item{}, err
	}

	item := external.Item{
		Source:      external.SourceReddit,
		ExternalID:  strings.TrimSpace(input.ID),
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.SelfText),
		URL:         normalizePostURL(input),
	}

	if err := validateNormalizedPost(item, input); err != nil {
		return external.Item{}, err
	}

	return item, nil
}

func normalizePostURL(input RedditPost) string {
	if strings.TrimSpace(input.URL) != "" {
		return strings.TrimSpace(input.URL)
	}

	permalink := strings.TrimSpace(input.Permalink)
	if permalink == "" {
		return ""
	}

	if strings.HasPrefix(permalink, "http://") ||
		strings.HasPrefix(permalink, "https://") {
		return permalink
	}

	return "https://www.reddit.com" +
		"/" +
		strings.TrimLeft(permalink, "/")
}

func validateNormalizedPost(
	item external.Item,
	input RedditPost,
) error {
	if err := item.Validate(); err != nil {
		return err
	}

	if item.Title == "" {
		return errors.New(
			"reddit post title is required",
		)
	}

	if item.URL != "" {
		parsed, err := url.Parse(item.URL)
		if err != nil {
			return err
		}

		if parsed.Scheme != "http" &&
			parsed.Scheme != "https" {
			return errors.New(
				"reddit post URL must use http or https",
			)
		}

		if parsed.Host == "" {
			return errors.New(
				"reddit post URL must include host",
			)
		}
	}

	if input.CreatedUTC < 0 {
		return errors.New(
			"reddit post created_utc must not be negative",
		)
	}

	return nil
}
