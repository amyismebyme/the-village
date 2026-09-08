package middleware

import (
	"net/http"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/amyismebyme/the-village/apps/api/internal/middleware"

// Tracing creates one server span for each inbound HTTP request and extracts
// any incoming W3C trace context. The route pattern is resolved by ServeMux
// while the wrapped handler executes, so the final normalized route is added
// to the span after the request is served.
func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if next == nil {
			return
		}

		tracer := otel.Tracer(tracerName)

		ctx := otel.GetTextMapPropagator().Extract(
			r.Context(),
			propagation.HeaderCarrier(r.Header),
		)

		start := time.Now()

		ctx, span := tracer.Start(
			ctx,
			"HTTP "+r.Method,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		r = r.WithContext(ctx)

		rec := httputil.NewResponseRecorder(w)
		next.ServeHTTP(rec, r)

		route := httputil.RouteLabel(r)
		status := rec.Status

		span.SetName(
			r.Method + " " + route,
		)

		span.SetAttributes(
			attribute.String(
				"http.request.method",
				r.Method,
			),
			attribute.String(
				"http.route",
				route,
			),
			attribute.Int(
				"http.response.status_code",
				status,
			),
			attribute.Int64(
				"http.server.duration_ms",
				time.Since(start).Milliseconds(),
			),
		)

		if status >= http.StatusInternalServerError {
			span.SetStatus(
				codes.Error,
				"HTTP server error",
			)
		}
	})
}
