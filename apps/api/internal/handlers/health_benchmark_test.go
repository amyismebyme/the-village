package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/health"
)

type benchmarkChecker struct{}

func (benchmarkChecker) Name() string {
	return "database"
}

func (benchmarkChecker) Check(context.Context) error {
	return nil
}

func BenchmarkHealthHandler(b *testing.B) {

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	registry := health.NewRegistry()

	registry.Register(
		benchmarkChecker{},
	)

	handler := NewHealthHandler(
		logger,
		registry,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		rec := httptest.NewRecorder()

		handler.ServeHTTP(
			rec,
			req,
		)

		if rec.Code != http.StatusOK {
			b.Fatalf(
				"unexpected status %d",
				rec.Code,
			)
		}
	}
}
