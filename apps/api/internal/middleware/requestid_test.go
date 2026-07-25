package middleware

import (
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

func TestRequestIDGeneratesRequestID(t *testing.T) {

	var requestID string

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID = GetRequestID(r.Context())
	}))

	req := testutil.NewRequest(http.MethodGet, "/")
	rr := testutil.NewRecorder()

	handler.ServeHTTP(rr, req)

	if requestID == "" {
		t.Fatal("expected request id")
	}

	if requestID == "unknown" {
		t.Fatal("expected generated request id")
	}

	headerID := rr.Header().Get("X-Request-ID")

	if headerID == "" {
		t.Fatal("expected X-Request-ID header")
	}

	if headerID != requestID {
		t.Fatal("request id in context and header should match")
	}
}

func TestGetRequestIDUnknown(t *testing.T) {

	id := GetRequestID(testutil.NewRequest(http.MethodGet, "/").Context())

	if id != "unknown" {
		t.Fatalf("expected unknown got %s", id)
	}
}
