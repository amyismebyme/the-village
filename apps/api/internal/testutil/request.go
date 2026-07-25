package testutil

import (
	"net/http"
	"net/http/httptest"
)

// NewRequest creates an HTTP request for tests.
func NewRequest(method, path string) *http.Request {
	return httptest.NewRequest(method, path, nil)
}

func NewRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}
