package worker

import (
	"context"
	"testing"
	"time"
)

type testWorker struct {
	run func(context.Context) error
}

func (w testWorker) Run(ctx context.Context) error {
	return w.run(ctx)
}

func TestWorkerContract(t *testing.T) {
	var worker Worker = testWorker{
		run: func(ctx context.Context) error {
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	if err := worker.Run(ctx); err != context.Canceled {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestWorkerContractRequiresContextCancellation(
	t *testing.T,
) {
	done := make(chan struct{})

	var worker Worker = testWorker{
		run: func(ctx context.Context) error {
			defer close(done)

			<-ctx.Done()

			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	go func() {
		_ = worker.Run(ctx)
	}()

	cancel()

	select {
	case <-done:
		return

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"worker did not stop after context cancellation",
		)
	}
}
