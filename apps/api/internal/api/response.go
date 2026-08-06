package api

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data any `json:"data,omitempty"`
}

func WriteJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(Response{
		Data: data,
	})
}
