package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes a JSON response with the supplied HTTP status code.
func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// At this point the status has already been written. There is
	// nothing useful the handler can do if encoding fails.
	_ = json.NewEncoder(w).Encode(data)
}

// writeNoContent sends a 204 response with no response body.
func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
