package testutil

import (
	"encoding/json"
	"testing"

	"net/http/httptest"
)

func DecodeJSON[T any](
	t *testing.T,
	rec *httptest.ResponseRecorder,
) T {

	t.Helper()

	var response T

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	); err != nil {

		t.Fatalf(
			"unable to decode JSON: %v",
			err,
		)

	}

	return response

}
