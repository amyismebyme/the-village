package health

import (
	"encoding/json"
	"net/http"
)

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// The response is already being written to the client.
	// There is nothing useful the handler can do if encoding fails
	// at this point, so intentionally ignore the error.
	_ = json.NewEncoder(w).Encode(value)
}
