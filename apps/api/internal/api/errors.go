package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

func WriteError(
	w http.ResponseWriter,
	status int,
	message string,
	details ...string,
) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(
		ErrorResponse{
			Error:   message,
			Details: details,
		},
	)
}

func BadRequest(
	w http.ResponseWriter,
	message string,
	details ...string,
) {

	WriteError(
		w,
		http.StatusBadRequest,
		message,
		details...,
	)
}

func NotFound(
	w http.ResponseWriter,
	message string,
) {

	WriteError(
		w,
		http.StatusNotFound,
		message,
	)
}

func Conflict(
	w http.ResponseWriter,
	message string,
) {

	WriteError(
		w,
		http.StatusConflict,
		message,
	)
}

func InternalServerError(
	w http.ResponseWriter,
) {

	WriteError(
		w,
		http.StatusInternalServerError,
		"internal server error",
	)
}
