package testutil

import (
	"net/http"
	"net/http/httptest"
)

func NewRequest(
	method string,
	path string,
) *http.Request {

	return httptest.NewRequest(
		method,
		path,
		nil,
	)

}
