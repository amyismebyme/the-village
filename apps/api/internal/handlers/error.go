package handlers

import (
	"errors"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	writeJSON(
		w,
		status,
		ErrorResponse{
			Error: ErrorDetail{
				Code:    code,
				Message: message,
			},
		},
	)
}

// writeCommunityServiceError translates service/repository errors
// into stable HTTP responses.
func writeCommunityServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrCommunityAlreadyExists):
		writeError(
			w,
			http.StatusConflict,
			"community_already_exists",
			"community with this slug already exists",
		)

	case errors.Is(err, service.ErrInvalidCommunity):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_community",
			"community validation failed",
		)

	case errors.Is(err, service.ErrNilCommunity):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			"community is required",
		)

	case errors.Is(err, repository.ErrNotFound):
		writeError(
			w,
			http.StatusNotFound,
			"community_not_found",
			"community not found",
		)

	case errors.Is(err, repository.ErrAlreadyExists):
		writeError(
			w,
			http.StatusConflict,
			"community_already_exists",
			"community with this slug already exists",
		)

	case errors.Is(err, service.ErrInvalidCommunityID):
		writeError(w, http.StatusBadRequest, "invalid community id", "invalid community id")

	default:
		writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"internal server error",
		)
	}
}
