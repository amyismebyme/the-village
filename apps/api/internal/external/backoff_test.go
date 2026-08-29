package external

import (
	"testing"
	"time"
)

func TestBackoffExponential(t *testing.T) {
	backoff, err := NewBackoff(
		time.Second,
		30*time.Second,
		2,
		0,
	)
	if err != nil {
		t.Fatalf(
			"create backoff: %v",
			err,
		)
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{
			attempt: 1,
			want:    time.Second,
		},
		{
			attempt: 2,
			want:    2 * time.Second,
		},
		{
			attempt: 3,
			want:    4 * time.Second,
		},
		{
			attempt: 4,
			want:    8 * time.Second,
		},
		{
			attempt: 5,
			want:    16 * time.Second,
		},
		{
			attempt: 6,
			want:    30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(
			time.Duration(tt.attempt).String(),
			func(t *testing.T) {
				got := backoff.Delay(tt.attempt)

				if got != tt.want {
					t.Fatalf(
						"expected %s, got %s",
						tt.want,
						got,
					)
				}
			},
		)
	}
}

func TestBackoffJitterRemainsBounded(
	t *testing.T,
) {
	backoff, err := NewBackoff(
		time.Second,
		10*time.Second,
		2,
		0.20,
	)
	if err != nil {
		t.Fatalf(
			"create backoff: %v",
			err,
		)
	}

	for range 100 {
		got := backoff.Delay(1)

		if got < 800*time.Millisecond ||
			got > 1200*time.Millisecond {
			t.Fatalf(
				"jitter outside expected range: %s",
				got,
			)
		}
	}
}
