package telemetry

import (
	"net/http"
	"strconv"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		tracer := otel.Tracer("github.com/amyismebyme/the-village/http")

		route := httputil.RouteLabel(r)
		ctx, span := tracer.Start(ctx, r.Method+" "+route, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		r = r.WithContext(ctx)
		rec := httputil.NewResponseRecorder(w)
		start := time.Now()
		next.ServeHTTP(rec, r)

		status := rec.Status
		span.SetAttributes(
			attribute.String("http.request.method", r.Method),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", status),
			attribute.Int64("http.server.duration_ms", time.Since(start).Milliseconds()),
		)
		if status >= 500 {
			span.SetStatus(codes.Error, "server error")
		} else {
			span.SetStatus(codes.Ok, strconv.Itoa(status))
		}
	})
}
