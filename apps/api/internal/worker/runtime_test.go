package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

type runtimeTestWorker struct {
	run func(context.Context) error
}

func (w *runtimeTestWorker) Run(
	ctx context.Context,
) error {
	return w.run(ctx)
}

func TestRuntimeIsolatesWorkerFailures(
	t *testing.T,
) {
	fatalErr := errors.New(
		"worker A failed",
	)

	workerAFinished := make(
		chan struct{},
	)

	workerBRunning := make(
		chan struct{},
	)

	workerBStopped := make(
		chan struct{},
	)

	workerA := &runtimeTestWorker{
		run: func(ctx context.Context) error {
			close(workerAFinished)

			return fatalErr
		},
	}

	workerB := &runtimeTestWorker{
		run: func(ctx context.Context) error {
			close(workerBRunning)

			<-ctx.Done()

			close(workerBStopped)

			return ctx.Err()
		},
	}

	runtime := NewRuntime()

	if err := runtime.Add(
		NewLifecycle(workerA),
	); err != nil {
		t.Fatalf(
			"add worker A: %v",
			err,
		)
	}

	if err := runtime.Add(
		NewLifecycle(workerB),
	); err != nil {
		t.Fatalf(
			"add worker B: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	if err := runtime.Start(ctx); err != nil {
		t.Fatalf(
			"start runtime: %v",
			err,
		)
	}

	select {
	case <-workerAFinished:

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"worker A did not execute",
		)
	}

	select {
	case <-workerBRunning:

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"worker B did not continue running after worker A failed",
		)
	}

	select {
	case err := <-runtime.Errors():
		if !errors.Is(
			err,
			fatalErr,
		) {
			t.Fatalf(
				"expected worker A error, got %v",
				err,
			)
		}

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"expected worker A error",
		)
	}

	cancel()

	select {
	case <-workerBStopped:

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"worker B did not stop after runtime cancellation",
		)
	}

	if err := runtime.Stop(
		context.Background(),
	); err != nil {
		t.Fatalf(
			"stop runtime: %v",
			err,
		)
	}
}
