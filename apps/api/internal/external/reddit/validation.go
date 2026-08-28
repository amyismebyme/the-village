package reddit

import (
	"strings"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func validateSubreddit(
	subreddit string,
) error {
	if strings.TrimSpace(subreddit) == "" {
		return external.ErrInvalidConfig
	}

	return nil
}
