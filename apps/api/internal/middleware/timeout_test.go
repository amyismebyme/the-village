package middleware

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestRequestTimeoutPropagatesDeadline(t *testing.T) {
	timeout := 20 * time.Millisecond

	var deadlineSeen bool

	handler := RequestTimeout(
		timeout,
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			deadline, ok := r.Context().Deadline()
			if !ok {
				t.Fatal("expected request deadline")
			}

			if time.Until(deadline) <= 0 ||
				time.Until(deadline) > timeout {
				t.Fatalf("unexpected request deadline: %v", deadline)
			}

			deadlineSeen = true
		}),
	)

	req := testutil.NewRequest(http.MethodGet, "/")
	rec := testutil.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !deadlineSeen {
		t.Fatal("expected request deadline to be propagated")
	}
}

func TestRequestTimeoutPreservesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var canceled bool

	handler := RequestTimeout(
		time.Second,
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if err := r.Context().Err(); err != context.Canceled {
				t.Fatalf("expected context canceled, got %v", err)
			}

			canceled = true
		}),
	)

	req := testutil.NewRequest(http.MethodGet, "/").WithContext(ctx)
	rec := testutil.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !canceled {
		t.Fatal("expected cancellation to propagate")
	}
}
