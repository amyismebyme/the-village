package reddit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func BenchmarkIngestionThroughputTelemetry(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"after":null,"before":null,"children":[
		{"kind":"t3","data":{"id":"bench-1","name":"t3_bench-1","title":"Benchmark One","selftext":"body","url":"https://example.test/1","permalink":"/r/toronto/comments/bench-1","subreddit":"toronto","author":"bench","created_utc":1234567890,"is_self":true}},
		{"kind":"t3","data":{"id":"bench-2","name":"t3_bench-2","title":"Benchmark Two","selftext":"body","url":"https://example.test/2","permalink":"/r/toronto/comments/bench-2","subreddit":"toronto","author":"bench","created_utc":1234567890,"is_self":true}}
	]}}`))
	}))
	defer server.Close()

	client, err := NewClient(
		server.Client(),
		server.URL,
		"the-village/benchmark",
		time.Second,
	)
	if err != nil {
		b.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		on   bool
	}{
		{name: "tracing_off", on: false},
		{name: "tracing_on", on: true},
	} {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			restore := installBenchmarkProvider(b, tc.on)
			defer restore()

			service := NewIngestionService(client, NewPostNormalizer())
			ctx := context.Background()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := service.IngestListing(
					ctx,
					"benchmark-token",
					"toronto",
					2,
					"",
				); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func installBenchmarkProvider(b *testing.B, enabled bool) func() {
	b.Helper()
	previous := otel.GetTracerProvider()

	if !enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())

		return func() {
			otel.SetTracerProvider(previous)
		}
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(provider)

	return func() {
		_ = provider.Shutdown(context.Background())
		otel.SetTracerProvider(previous)
	}
}
