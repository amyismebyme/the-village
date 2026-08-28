package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSchedulerRejectsInvalidInterval(t *testing.T) {
	_, err := NewScheduler(0)

	if err == nil {
		t.Fatal(
			"expected invalid interval error",
		)
	}
}

func TestSchedulerRunsImmediately(t *testing.T) {
	scheduler, err := NewScheduler(
		time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"create scheduler: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	runs := make(chan struct{}, 1)

	go func() {
		_ = scheduler.Run(
			ctx,
			func(context.Context) error {
				runs <- struct{}{}
				cancel()
				return nil
			},
		)
	}()

	select {
	case <-runs:
	case <-time.After(250 * time.Millisecond):
		t.Fatal(
			"expected immediate worker execution",
		)
	}
}

func TestSchedulerRepeats(t *testing.T) {
	scheduler, err := NewScheduler(
		10 * time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"create scheduler: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	runs := 0

	err = scheduler.Run(
		ctx,
		func(context.Context) error {
			runs++

			if runs >= 3 {
				cancel()
			}

			return nil
		},
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if runs != 3 {
		t.Fatalf(
			"expected 3 runs, got %d",
			runs,
		)
	}
}

func TestSchedulerCancellationInterruptsWait(
	t *testing.T,
) {
	scheduler, err := NewScheduler(
		time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"create scheduler: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	started := make(chan struct{})

	done := make(chan error, 1)

	go func() {
		done <- scheduler.Run(
			ctx,
			func(context.Context) error {
				close(started)
				return nil
			},
		)
	}()

	select {
	case <-started:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("scheduler did not run immediately")
	}

	cancel()

	select {
	case err := <-done:
		if !errors.Is(
			err,
			context.Canceled,
		) {
			t.Fatalf(
				"expected context.Canceled, got %v",
				err,
			)
		}

	case <-time.After(250 * time.Millisecond):
		t.Fatal(
			"scheduler did not stop after cancellation",
		)
	}
}

func TestSchedulerPropagatesFunctionError(
	t *testing.T,
) {
	scheduler, err := NewScheduler(
		time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"create scheduler: %v",
			err,
		)
	}

	expected := errors.New(
		"test worker failure",
	)

	err = scheduler.Run(
		context.Background(),
		func(context.Context) error {
			return expected
		},
	)

	if !errors.Is(err, expected) {
		t.Fatalf(
			"expected wrapped worker error, got %v",
			err,
		)
	}
}

func TestSchedulerResilientContinuesAfterFailure(
	t *testing.T,
) {
	scheduler, err := NewScheduler(
		10 * time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"create scheduler: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	expected := errors.New("temporary worker failure")

	var runs int
	var failures int

	err = scheduler.RunResilient(
		ctx,
		func(context.Context) error {
			runs++

			if runs == 1 {
				return expected
			}

			if runs >= 3 {
				cancel()
			}

			return nil
		},
		func(got error) {
			failures++

			if !errors.Is(got, expected) {
				t.Fatalf(
					"expected failure %v, got %v",
					expected,
					got,
				)
			}
		},
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if runs != 3 {
		t.Fatalf(
			"expected 3 runs, got %d",
			runs,
		)
	}

	if failures != 1 {
		t.Fatalf(
			"expected 1 failure, got %d",
			failures,
		)
	}
}

func TestSchedulerResilientDoesNotCountCancellationAsFailure(
	t *testing.T,
) {
	scheduler, err := NewScheduler(
		time.Hour,
	)
	if err != nil {
		t.Fatalf(
			"create scheduler: %v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	var failures int

	err = scheduler.RunResilient(
		ctx,
		func(ctx context.Context) error {
			cancel()
			return ctx.Err()
		},
		func(error) {
			failures++
		},
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}

	if failures != 0 {
		t.Fatalf(
			"expected zero worker failures during cancellation, got %d",
			failures,
		)
	}
}
