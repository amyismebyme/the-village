package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
	"github.com/amyismebyme/the-village/apps/api/internal/validation"
)

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeError(
		recorder,
		http.StatusBadRequest,
		"invalid_request",
		"request body contains invalid JSON",
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			contentType,
		)
	}

	var response errorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if response.Error.Code != "invalid_request" {
		t.Fatalf(
			"expected code %q, got %q",
			"invalid_request",
			response.Error.Code,
		)
	}

	if response.Error.Message != "request body contains invalid JSON" {
		t.Fatalf(
			"unexpected message: %q",
			response.Error.Message,
		)
	}
}

func TestWriteCommunityServiceError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "invalid community id",
			err:            service.ErrInvalidCommunityID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_id",
		},
		{
			name:           "invalid community",
			err:            service.ErrInvalidCommunity,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_community",
		},
		{
			name:           "duplicate community",
			err:            service.ErrCommunityAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedCode:   "community_already_exists",
		},
		{
			name:           "not found",
			err:            repository.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "community_not_found",
		},
		{
			name:           "database failure",
			err:            errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "internal_error",
		},
		{
			name: "request timeout",
			err: fmt.Errorf(
				"database: %w",
				context.DeadlineExceeded,
			),
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   "request_timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			gotStatus := writeServiceError(
				recorder,
				tt.err,
			)

			if gotStatus != tt.expectedStatus {
				t.Fatalf(
					"expected returned status %d, got %d",
					tt.expectedStatus,
					gotStatus,
				)
			}

			if recorder.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}

			var response errorResponse

			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf(
					"decode response: %v",
					err,
				)
			}

			if response.Error.Code != tt.expectedCode {
				t.Fatalf(
					"expected code %q, got %q",
					tt.expectedCode,
					response.Error.Code,
				)
			}
		})
	}
}

func TestWriteCommunityServiceErrorTimeout(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := fmt.Errorf(
		"community service: database operation: %w",
		context.DeadlineExceeded,
	)

	writeServiceError(
		recorder,
		err,
	)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusGatewayTimeout,
			recorder.Code,
		)
	}

	var response errorResponse

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if response.Error.Code != "request_timeout" {
		t.Fatalf(
			"expected code %q, got %q",
			"request_timeout",
			response.Error.Code,
		)
	}
}

func TestWriteCommunityServiceErrorPreservesValidationDetail(
	t *testing.T,
) {
	recorder := httptest.NewRecorder()

	validationErr := fmt.Errorf(
		"name: %w",
		validation.ErrTooShort,
	)

	err := fmt.Errorf(
		"%w: %w",
		service.ErrInvalidCommunity,
		validationErr,
	)

	writeServiceError(
		recorder,
		err,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	var response errorResponse

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if response.Error.Code != "invalid_community" {
		t.Fatalf(
			"expected code %q, got %q",
			"invalid_community",
			response.Error.Code,
		)
	}

	if response.Error.Message != "name: value is too short" {
		t.Fatalf(
			"expected validation detail %q, got %q",
			"name: value is too short",
			response.Error.Message,
		)
	}
}
