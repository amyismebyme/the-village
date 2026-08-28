package worker

import (
	"context"
	"testing"
	"time"
)

func TestNewRunContextPreservesParentCancellation(
	t *testing.T,
) {
	parent, cancel := context.WithCancel(
		context.Background(),
	)

	runCtx := NewRunContext(parent)

	cancel()

	select {
	case <-runCtx.Done():
	case <-time.After(250 * time.Millisecond):
		t.Fatal(
			"worker context did not inherit cancellation",
		)
	}

	if runCtx.Err() != context.Canceled {
		t.Fatalf(
			"expected context.Canceled, got %v",
			runCtx.Err(),
		)
	}
}

func TestNewRunContextPreservesParentDeadline(
	t *testing.T,
) {
	parent, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer cancel()

	runCtx := NewRunContext(parent)

	parentDeadline, ok := parent.Deadline()
	if !ok {
		t.Fatal("expected parent deadline")
	}

	runDeadline, ok := runCtx.Deadline()
	if !ok {
		t.Fatal("expected worker deadline")
	}

	if runDeadline.After(parentDeadline) {
		t.Fatalf(
			"worker deadline %v exceeds parent deadline %v",
			runDeadline,
			parentDeadline,
		)
	}
}
