package reddit

import (
	"errors"
	"strings"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func IdentityForPost(
	post RedditPost,
) (external.Identity, error) {
	externalID := strings.TrimSpace(post.ID)

	if externalID == "" {
		return external.Identity{}, errors.New(
			"reddit post ID is required",
		)
	}

	identity := external.Identity{
		Source:     external.SourceReddit,
		ExternalID: externalID,
	}

	if err := identity.Validate(); err != nil {
		return external.Identity{}, err
	}

	return identity, nil
}
