package telemetry

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func BenchmarkHTTPMiddlewareTelemetry(b *testing.B) {
	for _, tc := range []struct {
		name string
		on   bool
	}{
		{name: "telemetry_off", on: false},
		{name: "telemetry_on", on: true},
	} {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			restore := installBenchmarkProvider(b, tc.on)
			defer restore()

			handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, req)
			}
		})
	}
}

func installBenchmarkProvider(b *testing.B, enabled bool) func() {
	b.Helper()
	previous := otel.GetTracerProvider()

	if !enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return func() { otel.SetTracerProvider(previous) }
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(provider)

	return func() {
		_ = provider.Shutdown(b.Context())
		otel.SetTracerProvider(previous)
	}
}
