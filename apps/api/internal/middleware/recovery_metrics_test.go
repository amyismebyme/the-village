package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecoveryIncrementsPanicMetric(t *testing.T) {
	before := testutil.ToFloat64(metrics.PanicsTotal)

	handler := Recovery(nil, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	after := testutil.ToFloat64(metrics.PanicsTotal)
	if got := after - before; got != 1 {
		t.Fatalf("expected panic metric +1, got %v", got)
	}
}
