package testutil

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestControlledWorkerTracksRuns(t *testing.T) {
	expected := errors.New("test failure")

	worker := NewControlledWorker(
		func(context.Context) error {
			return expected
		},
	)

	err := worker.Run(
		context.Background(),
	)

	if !errors.Is(err, expected) {
		t.Fatalf(
			"expected %v, got %v",
			expected,
			err,
		)
	}

	if worker.RunCount() != 1 {
		t.Fatalf(
			"expected one run, got %d",
			worker.RunCount(),
		)
	}

	select {
	case <-worker.Started:

	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected Started signal")
	}

	select {
	case <-worker.Stopped:

	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected Stopped signal")
	}
}
