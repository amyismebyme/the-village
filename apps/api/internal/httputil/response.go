package httputil

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		// At this point headers have already been written, so there is
		// no useful HTTP response we can send. The caller's logger/
		// middleware can observe the completed response.
		return
	}
}
