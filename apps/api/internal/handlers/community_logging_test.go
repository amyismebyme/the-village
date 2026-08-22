package handlers

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/middleware"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

func TestCreateCommunityLogsStructuredOperation(
	t *testing.T,
) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	handler := NewHandler(
		&communityServiceMock{
			createFunc: func(
				_ context.Context,
				community *model.Community,
			) error {
				community.ID = 42
				return nil
			},
		},
		logger,
	)

	req := newCommunityRequest(
		http.MethodPost,
		"/api/v1/communities",
		`{
			"name": "Toronto Men",
			"slug": "toronto-men"
		}`,
	)

	req = req.WithContext(
		middleware.WithRequestID(
			req.Context(),
			"create-request-id",
		),
	)

	recorder := httptest.NewRecorder()

	handler.CreateCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	output := logs.String()

	for _, expected := range []string{
		`msg="community operation completed"`,
		"request_id=create-request-id",
		"operation=create",
		"community_id=42",
		"status=201",
		"duration_ms=",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected log to contain %q; got:\n%s",
				expected,
				output,
			)
		}
	}
}

func TestUpdateCommunityLogsStructuredOperation(
	t *testing.T,
) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	handler := NewHandler(
		&communityServiceMock{},
		logger,
	)

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/42",
		`{
			"name": "Toronto Men",
			"slug": "toronto-men"
		}`,
	)

	req.SetPathValue("id", "42")

	req = req.WithContext(
		middleware.WithRequestID(
			req.Context(),
			"update-request-id",
		),
	)

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	output := logs.String()

	for _, expected := range []string{
		"request_id=update-request-id",
		"operation=update",
		"community_id=42",
		"status=200",
		"duration_ms=",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected log to contain %q; got:\n%s",
				expected,
				output,
			)
		}
	}
}

func TestDeleteCommunityLogsStructuredOperation(
	t *testing.T,
) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	handler := NewHandler(
		&communityServiceMock{},
		logger,
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/communities/42",
		nil,
	)

	req.SetPathValue("id", "42")

	req = req.WithContext(
		middleware.WithRequestID(
			req.Context(),
			"delete-request-id",
		),
	)

	recorder := httptest.NewRecorder()

	handler.DeleteCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			recorder.Code,
		)
	}

	output := logs.String()

	for _, expected := range []string{
		"request_id=delete-request-id",
		"operation=delete",
		"community_id=42",
		"status=204",
		"duration_ms=",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected log to contain %q; got:\n%s",
				expected,
				output,
			)
		}
	}
}
