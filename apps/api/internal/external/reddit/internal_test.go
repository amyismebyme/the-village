package reddit

import (
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
)

func TestIdentityForPost(t *testing.T) {
	identity, err := IdentityForPost(
		RedditPost{
			ID: " abc123 ",
		},
	)
	if err != nil {
		t.Fatalf(
			"create identity: %v",
			err,
		)
	}

	if identity.Source != external.SourceReddit {
		t.Fatalf(
			"unexpected source %q",
			identity.Source,
		)
	}

	if identity.ExternalID != "abc123" {
		t.Fatalf(
			"unexpected external ID %q",
			identity.ExternalID,
		)
	}

	if identity.Key() != "reddit:abc123" {
		t.Fatalf(
			"unexpected identity key %q",
			identity.Key(),
		)
	}
}

func TestIdentityForPostRequiresID(t *testing.T) {
	_, err := IdentityForPost(
		RedditPost{},
	)

	if err == nil {
		t.Fatal(
			"expected missing Reddit post ID error",
		)
	}
}
