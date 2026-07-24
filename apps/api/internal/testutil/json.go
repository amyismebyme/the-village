package testutil

import (
	"encoding/json"
	"testing"
)

// DecodeJSON decodes a JSON response into the provided destination.
func DecodeJSON[T any](t *testing.T, data []byte, dst *T) {
	t.Helper()

	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
}