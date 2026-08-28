package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

type lifecycleWorker struct {
	run func(context.Context) error
}

func (w *lifecycleWorker) Run(
	ctx context.Context,
) error {
	return w.run(ctx)
}

func TestLifecycleRunsWorker(t *testing.T) {
	called := false

	w := &lifecycleWorker{
		run: func(ctx context.Context) error {
			called = true
			return nil
		},
	}

	lifecycle := NewLifecycle(w)

	if err := lifecycle.Run(
		context.Background(),
	); err != nil {
		t.Fatalf(
			"expected nil, got %v",
			err,
		)
	}

	if !called {
		t.Fatal(
			"expected worker Run to be called",
		)
	}

	if lifecycle.Running() {
		t.Fatal(
			"expected lifecycle to stop after Run returns",
		)
	}
}

func TestLifecyclePropagatesWorkerError(
	t *testing.T,
) {
	expected := errors.New("worker failed")

	w := &lifecycleWorker{
		run: func(ctx context.Context) error {
			return expected
		},
	}

	lifecycle := NewLifecycle(w)

	err := lifecycle.Run(
		context.Background(),
	)
	if err == nil {
		t.Fatal(
			"expected worker error",
		)
	}

	if !errors.Is(err, expected) {
		t.Fatalf(
			"expected wrapped worker error, got %v",
			err,
		)
	}
}

func TestLifecycleRejectsCancelledContext(
	t *testing.T,
) {
	called := false

	w := &lifecycleWorker{
		run: func(ctx context.Context) error {
			called = true
			return nil
		},
	}

	lifecycle := NewLifecycle(w)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	err := lifecycle.Run(ctx)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if called {
		t.Fatal(
			"worker should not run with cancelled context",
		)
	}
}

func TestLifecycleStopsAfterWorkerHonorsCancellation(
	t *testing.T,
) {
	started := make(chan struct{})
	stopped := make(chan struct{})

	w := &lifecycleWorker{
		run: func(ctx context.Context) error {
			close(started)

			<-ctx.Done()

			close(stopped)

			return ctx.Err()
		},
	}

	lifecycle := NewLifecycle(w)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	done := make(chan error, 1)

	go func() {
		done <- lifecycle.Run(ctx)
	}()

	select {
	case <-started:
	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal("worker did not start")
	}

	if !lifecycle.Running() {
		t.Fatal(
			"expected lifecycle to report running",
		)
	}

	cancel()

	select {
	case <-stopped:
	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"worker did not stop after cancellation",
		)
	}

	err := <-done

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if lifecycle.Running() {
		t.Fatal(
			"expected lifecycle to report stopped",
		)
	}
}

func TestLifecycleRejectsConcurrentRun(
	t *testing.T,
) {
	started := make(chan struct{})
	release := make(chan struct{})

	w := &lifecycleWorker{
		run: func(ctx context.Context) error {
			close(started)

			select {
			case <-release:
				return nil

			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	lifecycle := NewLifecycle(w)

	firstDone := make(chan error, 1)

	go func() {
		firstDone <- lifecycle.Run(
			context.Background(),
		)
	}()

	select {
	case <-started:
	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal("worker did not start")
	}

	err := lifecycle.Run(
		context.Background(),
	)

	if err == nil {
		t.Fatal(
			"expected concurrent Run to be rejected",
		)
	}

	close(release)

	if err := <-firstDone; err != nil {
		t.Fatalf(
			"first Run returned error: %v",
			err,
		)
	}
}

func TestLifecycleStopGracefullyCancelsWorker(
	t *testing.T,
) {
	started := make(chan struct{})
	stopped := make(chan struct{})

	w := &lifecycleWorker{
		run: func(ctx context.Context) error {
			close(started)

			<-ctx.Done()

			close(stopped)

			return nil
		},
	}

	lifecycle := NewLifecycle(w)

	runDone := make(chan error, 1)

	go func() {
		runDone <- lifecycle.Run(
			context.Background(),
		)
	}()

	select {
	case <-started:

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal("worker did not start")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	if err := lifecycle.Stop(
		shutdownCtx,
	); err != nil {
		t.Fatalf(
			"graceful stop failed: %v",
			err,
		)
	}

	select {
	case <-stopped:

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal(
			"worker did not receive cancellation",
		)
	}

	if err := <-runDone; err != nil {
		t.Fatalf(
			"worker Run returned error: %v",
			err,
		)
	}

	if lifecycle.Running() {
		t.Fatal(
			"expected lifecycle to report stopped",
		)
	}
}

func TestLifecycleStopWhenNotRunning(
	t *testing.T,
) {
	lifecycle := NewLifecycle(
		&lifecycleWorker{
			run: func(context.Context) error {
				return nil
			},
		},
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	if err := lifecycle.Stop(ctx); err != nil {
		t.Fatalf(
			"expected nil, got %v",
			err,
		)
	}
}

type blockingLifecycleWorker struct {
	started chan struct{}
	release chan struct{}
}

func (w *blockingLifecycleWorker) Run(
	ctx context.Context,
) error {
	close(w.started)

	<-w.release

	return nil
}

func TestLifecycleStopHonorsTimeout(
	t *testing.T,
) {
	worker := &blockingLifecycleWorker{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}

	lifecycle := NewLifecycle(worker)

	go func() {
		_ = lifecycle.Run(
			context.Background(),
		)
	}()

	select {
	case <-worker.started:

	case <-time.After(
		250 * time.Millisecond,
	):
		t.Fatal("worker did not start")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		25*time.Millisecond,
	)
	defer cancel()

	err := lifecycle.Stop(ctx)

	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"expected context.DeadlineExceeded, got %v",
			err,
		)
	}

	close(worker.release)
}
