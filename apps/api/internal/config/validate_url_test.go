package config

import (
	"strings"
	"testing"
)

func TestValidateHTTPSURLUsesProvidedNameInError(t *testing.T) {
	err := validateHTTPSURL(
		"REDDIT_AUTH_BASE_URL",
		"not-a-url",
	)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "REDDIT_AUTH_BASE_URL") {
		t.Fatalf(
			"expected error to mention provided name, got %v",
			err,
		)
	}

	if strings.Contains(err.Error(), "REDDIT_BASE_URL is invalid") {
		t.Fatalf(
			"error still contains hard-coded Reddit base URL name: %v",
			err,
		)
	}
}
