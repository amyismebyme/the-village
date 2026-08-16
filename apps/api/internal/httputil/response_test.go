package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(
		rec,
		http.StatusOK,
		map[string]string{
			"status": "healthy",
		},
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"expected application/json content type, got %q",
			got,
		)
	}

	if rec.Body.Len() == 0 {
		t.Fatal("expected JSON response body")
	}
}

func TestResponseRecorderCapturesStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseRecorder(rec)

	w.WriteHeader(http.StatusNotFound)

	if w.Status != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			w.Status,
		)
	}
}

func TestResponseRecorderDefaultsTo200(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseRecorder(rec)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if w.Status != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			w.Status,
		)
	}
}

func TestResponseRecorderIgnoresSubsequentWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()

	w := NewResponseRecorder(rec)

	w.WriteHeader(http.StatusCreated)
	w.WriteHeader(http.StatusInternalServerError)

	if w.Status != http.StatusCreated {
		t.Fatalf(
			"expected original status %d, got %d",
			http.StatusCreated,
			w.Status,
		)
	}
}
