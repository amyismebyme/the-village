package external

import (
	"context"
	"testing"
	"time"
)

func TestWithRequestTimeoutPropagatesCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())

	ctx, cleanup := WithRequestTimeout(
		parent,
		time.Second,
	)
	defer cleanup()

	cancel()

	select {
	case <-ctx.Done():
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected cancellation to propagate")
	}

	if ctx.Err() != context.Canceled {
		t.Fatalf(
			"expected context.Canceled, got %v",
			ctx.Err(),
		)
	}
}

func TestWithRequestTimeoutDoesNotExtendParentDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer cancel()

	ctx, cleanup := WithRequestTimeout(
		parent,
		time.Second,
	)
	defer cleanup()

	parentDeadline, ok := parent.Deadline()
	if !ok {
		t.Fatal("expected parent deadline")
	}

	childDeadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected child deadline")
	}

	if childDeadline.After(parentDeadline) {
		t.Fatalf(
			"child deadline %v exceeds parent deadline %v",
			childDeadline,
			parentDeadline,
		)
	}
}

func TestWithRequestTimeoutUsesConfiguredTimeout(t *testing.T) {
	parent := context.Background()

	ctx, cancel := WithRequestTimeout(
		parent,
		50*time.Millisecond,
	)
	defer cancel()

	start := time.Now()

	<-ctx.Done()

	if ctx.Err() != context.DeadlineExceeded {
		t.Fatalf(
			"expected context deadline exceeded, got %v",
			ctx.Err(),
		)
	}

	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Fatalf(
			"timeout took too long: %s",
			elapsed,
		)
	}
}
