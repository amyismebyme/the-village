package worker

import (
	"context"
)

// RunContext represents the context owned by a worker invocation.
//
// Workers should use this context for all downstream operations:
// external requests, database calls, timers, and child operations.
type RunContext struct {
	context.Context
}

// NewRunContext creates a worker execution context.
//
// The supplied parent remains authoritative. In particular, this does
// not create a detached background context.
func NewRunContext(
	parent context.Context,
) RunContext {
	if parent == nil {
		parent = context.Background()
	}

	return RunContext{
		Context: parent,
	}
}
