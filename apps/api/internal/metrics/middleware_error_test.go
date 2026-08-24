package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMiddlewareIncrementsErrorMetricForClientError(t *testing.T) {
	before := testutil.ToFloat64(ErrorsTotal.WithLabelValues("client"))

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(recorder, request)

	after := testutil.ToFloat64(ErrorsTotal.WithLabelValues("client"))
	if got := after - before; got != 1 {
		t.Fatalf("expected client error metric +1, got %v", got)
	}
}

func TestMiddlewareIncrementsErrorMetricForServerError(t *testing.T) {
	before := testutil.ToFloat64(ErrorsTotal.WithLabelValues("server"))

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(recorder, request)

	after := testutil.ToFloat64(ErrorsTotal.WithLabelValues("server"))
	if got := after - before; got != 1 {
		t.Fatalf("expected server error metric +1, got %v", got)
	}
}
