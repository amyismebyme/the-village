package handlers

import (
	"errors"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	response := errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	}

	httputil.WriteJSON(
		w,
		status,
		response,
	)
}

func writeServiceError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, service.ErrInvalidCommunityID):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"invalid community id",
		)

	case errors.Is(err, service.ErrInvalidCommunity):
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_community",
			"invalid community",
		)

	case errors.Is(err, service.ErrCommunityAlreadyExists):
		writeError(
			w,
			http.StatusConflict,
			"community_already_exists",
			"community already exists",
		)

	case errors.Is(err, repository.ErrNotFound):
		writeError(
			w,
			http.StatusNotFound,
			"community_not_found",
			"community not found",
		)

	default:
		writeError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"internal server error",
		)
	}
}
