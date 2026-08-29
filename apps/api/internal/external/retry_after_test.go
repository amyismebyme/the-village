package external

import (
	"net/http"
	"testing"
	"time"
)

func TestParseRetryAfterSeconds(t *testing.T) {
	got := ParseRetryAfter(
		"7",
		time.Now(),
	)

	if got != 7*time.Second {
		t.Fatalf(
			"expected 7s, got %s",
			got,
		)
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	now := time.Date(
		2026,
		8,
		29,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	target := now.Add(
		10 * time.Second,
	).UTC().Format(
		http.TimeFormat,
	)

	got := ParseRetryAfter(
		target,
		now,
	)

	if got < 9*time.Second ||
		got > 10*time.Second {
		t.Fatalf(
			"expected approximately 10s, got %s",
			got,
		)
	}
}

func TestParseRetryAfterInvalid(t *testing.T) {
	got := ParseRetryAfter(
		"not-a-delay",
		time.Now(),
	)

	if got != 0 {
		t.Fatalf(
			"expected zero duration, got %s",
			got,
		)
	}
}
