package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
	"github.com/amyismebyme/the-village/apps/api/internal/validation"
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
	case errors.Is(err, context.DeadlineExceeded):
		writeError(
			w,
			http.StatusGatewayTimeout,
			"request_timeout",
			"request timed out while processing the request",
		)

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
			validationMessage(err),
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

func validationMessage(err error) string {
	var multi interface {
		Unwrap() []error
	}

	if !errors.As(err, &multi) {
		return "invalid community"
	}

	for _, cause := range multi.Unwrap() {
		switch {
		case errors.Is(cause, validation.ErrRequired),
			errors.Is(cause, validation.ErrTooShort),
			errors.Is(cause, validation.ErrTooLong),
			errors.Is(cause, validation.ErrInvalidSlug):
			return cause.Error()
		}
	}

	return "invalid community"
}