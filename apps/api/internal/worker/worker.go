package worker

import "context"

// Worker represents a long-running background component.
//
// Run must block until the worker stops or its context is cancelled.
// Implementations must honor ctx.Done() and must not create an
// independent background context for normal execution.
type Worker interface {
	Run(ctx context.Context) error
}
