package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			writeCommunityServiceError(
				recorder,
				tt.err,
			)

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
