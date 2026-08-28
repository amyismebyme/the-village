package testutil

import (
	"context"
	"sync"
)

type ControlledWorker struct {
	mu sync.Mutex

	Runs int

	RunFunc func(context.Context) error

	Started chan struct{}
	Stopped chan struct{}
}

func NewControlledWorker(
	runFunc func(context.Context) error,
) *ControlledWorker {
	return &ControlledWorker{
		RunFunc: runFunc,
		Started: make(chan struct{}, 100),
		Stopped: make(chan struct{}, 100),
	}
}

func (w *ControlledWorker) Run(
	ctx context.Context,
) error {
	w.mu.Lock()
	w.Runs++
	w.mu.Unlock()

	w.Started <- struct{}{}

	if w.RunFunc == nil {
		<-ctx.Done()

		w.Stopped <- struct{}{}

		return ctx.Err()
	}

	err := w.RunFunc(ctx)

	w.Stopped <- struct{}{}

	return err
}

func (w *ControlledWorker) RunCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.Runs
}

type SequenceWorker struct {
	mu sync.Mutex

	results []error
	index   int

	Started chan struct{}
	Stopped chan struct{}
}

func NewSequenceWorker(
	results ...error,
) *SequenceWorker {
	return &SequenceWorker{
		results: results,
		Started: make(chan struct{}, 100),
		Stopped: make(chan struct{}, 100),
	}
}

func (w *SequenceWorker) Run(
	ctx context.Context,
) error {
	w.Started <- struct{}{}

	w.mu.Lock()

	defer w.mu.Unlock()

	var err error

	if w.index < len(w.results) {
		err = w.results[w.index]
		w.index++
	}

	w.Stopped <- struct{}{}

	return err
}
