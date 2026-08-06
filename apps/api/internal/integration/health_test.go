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
	ctx context.Context,
	community *model.Community,
) error {
	return nil
}

func (stubCommunityService) Get(
	ctx context.Context,
	id int64,
) (*model.Community, error) {
	return nil, nil
}

func (stubCommunityService) List(
	ctx context.Context,
) ([]*model.Community, error) {
	return nil, nil
}

func (stubCommunityService) Update(
	ctx context.Context,
	community *model.Community,
) error {
	return nil
}

func (stubCommunityService) Delete(
	ctx context.Context,
	id int64,
) error {
	return nil
}

func TestHealthEndpoint(t *testing.T) {
	db := OpenTestDatabase(t)

	registry := health.NewRegistry()
	registry.Register(database.NewHealthChecker(db))

	appLogger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	handler := handlers.NewHandler(stubCommunityService{})

	testServer := httptest.NewServer(
		server.NewRouter(appLogger, registry, handler),
	)
	t.Cleanup(testServer.Close)

	response, err := testServer.Client().Get(
		testServer.URL + "/health",
	)
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
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

	var body handlers.HealthResponse

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf(
			"decode health response: %v",
			err,
		)
	}

	if body.Status != "healthy" {
		t.Fatalf(
			"expected healthy, got %q",
			body.Status,
		)
	}

	databaseStatus, exists := body.Checks[database.CheckerName]
	if !exists {
		t.Fatal("database health check missing")
	}

	if databaseStatus != "healthy" {
		t.Fatalf(
			"expected database check to be healthy, got %q",
			databaseStatus,
		)
	}
}
