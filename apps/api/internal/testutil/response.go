package testutil

import "net/http/httptest"

func NewRecorder() *httptest.ResponseRecorder {

	return httptest.NewRecorder()

}
