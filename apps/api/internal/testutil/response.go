package testutil

import (
	"io"
	"log/slog"
)

// NewDiscardLogger returns a logger that discards all output.
func NewDiscardLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)
}
