//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
)

type stubCommunityService struct{}

func (stubCommunityService) Create(
	_ context.Context,
	_ *model.Community,
) error {
	return nil
}

func (stubCommunityService) Get(
	_ context.Context,
	_ int64,
) (*model.Community, error) {
	return nil, nil
}

func (stubCommunityService) List(
	_ context.Context,
) ([]*model.Community, error) {
	return nil, nil
}

func (stubCommunityService) Update(
	_ context.Context,
	_ *model.Community,
) error {
	return nil
}

func (stubCommunityService) Delete(
	_ context.Context,
	_ int64,
) error {
	return nil
}

func TestHealthEndpoint(t *testing.T) {
	db := OpenTestDatabase(t)

	registry := health.NewRegistry()
	registry.Register(
		database.NewHealthChecker(db),
	)

	appLogger := slog.New(
		slog.NewTextHandler(
			io.Discard,
			nil,
		),
	)

	handler := handlers.NewHandler(
		stubCommunityService{},
	)

	testServer := httptest.NewServer(
		server.NewRouter(
			appLogger,
			registry,
			handler,
		),
	)

	t.Cleanup(testServer.Close)

	t.Run("liveness", func(t *testing.T) {
		response, err := testServer.Client().Get(
			testServer.URL + "/health",
		)
		if err != nil {
			t.Fatalf(
				"GET /health failed: %v",
				err,
			)
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			t.Fatalf(
				"expected /health status %d, got %d",
				http.StatusOK,
				response.StatusCode,
			)
		}

		if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
			t.Fatalf(
				"expected application/json, got %q",
				contentType,
			)
		}

		var body health.HealthResponse

		if err := json.NewDecoder(
			response.Body,
		).Decode(&body); err != nil {
			t.Fatalf(
				"decode /health response: %v",
				err,
			)
		}

		if body.Status != "healthy" {
			t.Fatalf(
				"expected /health status %q, got %q",
				"healthy",
				body.Status,
			)
		}

		// Liveness must not expose dependency checks.
		if len(body.Checks) != 0 {
			t.Fatalf(
				"expected /health to contain no dependency checks, got %+v",
				body.Checks,
			)
		}
	})

	t.Run("readiness", func(t *testing.T) {
		response, err := testServer.Client().Get(
			testServer.URL + "/ready",
		)
		if err != nil {
			t.Fatalf(
				"GET /ready failed: %v",
				err,
			)
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			t.Fatalf(
				"expected /ready status %d, got %d",
				http.StatusOK,
				response.StatusCode,
			)
		}

		if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
			t.Fatalf(
				"expected application/json, got %q",
				contentType,
			)
		}

		var body health.HealthResponse

		if err := json.NewDecoder(
			response.Body,
		).Decode(&body); err != nil {
			t.Fatalf(
				"decode /ready response: %v",
				err,
			)
		}

		if body.Status != "ready" {
			t.Fatalf(
				"expected /ready status %q, got %q",
				"ready",
				body.Status,
			)
		}

		var databaseCheck *health.Result

		for i := range body.Checks {
			if body.Checks[i].Name == database.CheckerName {
				databaseCheck = &body.Checks[i]
				break
			}
		}

		if databaseCheck == nil {
			t.Fatalf(
				"database readiness check missing; expected %q, got %+v",
				database.CheckerName,
				body.Checks,
			)
		}

		if databaseCheck.Error != "" {
			t.Fatalf(
				"expected database readiness check to be healthy, got error %q",
				databaseCheck.Error,
			)
		}
	})
}
