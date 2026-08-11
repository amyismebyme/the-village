//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/logger"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

// -----------------------------------------------------------------------------
// Test application
// -----------------------------------------------------------------------------

type communityAPITestApp struct {
	server *httptest.Server
	repo   *postgres.CommunityRepository
}

func newCommunityAPITestApp(t *testing.T) *communityAPITestApp {
	t.Helper()

	cfg := config.Load()

	db, err := database.Open(
		context.Background(),
		cfg.Database,
	)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := postgres.NewCommunityRepository(
		db.Pool(),
	)

	// Clean database before test.
	if err := repo.DeleteAll(context.Background()); err != nil {
		t.Fatalf(
			"clean communities before test: %v",
			err,
		)
	}

	// Clean database after test.
	t.Cleanup(func() {
		if err := repo.DeleteAll(context.Background()); err != nil {
			t.Logf(
				"clean communities after test: %v",
				err,
			)
		}
	})

	communityService := service.NewCommunityService(
		repo,
	)

	handler := handlers.NewHandler(
		communityService,
	)

	appLogger := logger.New(cfg)

	healthRegistry := health.NewRegistry()

	httpHandler := server.NewRouter(
		appLogger,
		healthRegistry,
		handler,
	)

	testServer := httptest.NewServer(
		httpHandler,
	)

	t.Cleanup(func() {
		testServer.Close()
	})

	return &communityAPITestApp{
		server: testServer,
		repo:   repo,
	}
}

// -----------------------------------------------------------------------------
// HTTP helper
// -----------------------------------------------------------------------------

func communityAPIRequest(
	t *testing.T,
	app *communityAPITestApp,
	method string,
	path string,
	body string,
) *http.Response {
	t.Helper()

	var requestBody *strings.Reader

	if body == "" {
		requestBody = strings.NewReader("")
	} else {
		requestBody = strings.NewReader(body)
	}

	req, err := http.NewRequest(
		method,
		app.server.URL+path,
		requestBody,
	)
	if err != nil {
		t.Fatalf(
			"create request: %v",
			err,
		)
	}

	if body != "" {
		req.Header.Set(
			"Content-Type",
			"application/json",
		)
	}

	resp, err := app.server.Client().Do(req)
	if err != nil {
		t.Fatalf(
			"execute %s %s: %v",
			method,
			path,
			err,
		)
	}

	return resp
}

// -----------------------------------------------------------------------------
// Response helpers
// -----------------------------------------------------------------------------

func decodeCommunity(
	t *testing.T,
	resp *http.Response,
) model.Community {
	t.Helper()

	defer resp.Body.Close()

	var community model.Community

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&community); err != nil {
		t.Fatalf(
			"decode community response: %v",
			err,
		)
	}

	return community
}

type communityListResponse struct {
	Communities []*model.Community `json:"communities"`
}

func decodeCommunityList(
	t *testing.T,
	resp *http.Response,
) communityListResponse {
	t.Helper()

	defer resp.Body.Close()

	var response communityListResponse

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode community list response: %v",
			err,
		)
	}

	return response
}

func communityIDPath(id int64) string {
	return "/api/v1/communities/" +
		strconv.FormatInt(id, 10)
}

// -----------------------------------------------------------------------------
// POST → GET
// -----------------------------------------------------------------------------

func TestCommunityAPI_CreateThenGet(t *testing.T) {
	app := newCommunityAPITestApp(t)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men-integration",
		"description": "A community for men in Toronto.",
		"external_source": "manual"
	}`

	createResp := communityAPIRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	if createResp.StatusCode != http.StatusCreated {
		defer createResp.Body.Close()

		t.Fatalf(
			"expected POST status %d, got %d",
			http.StatusCreated,
			createResp.StatusCode,
		)
	}

	created := decodeCommunity(
		t,
		createResp,
	)

	if created.ID <= 0 {
		t.Fatalf(
			"expected created community ID > 0, got %d",
			created.ID,
		)
	}

	if created.Name != "Toronto Men" {
		t.Fatalf(
			"expected name %q, got %q",
			"Toronto Men",
			created.Name,
		)
	}

	if created.Slug != "toronto-men-integration" {
		t.Fatalf(
			"expected slug %q, got %q",
			"toronto-men-integration",
			created.Slug,
		)
	}

	if created.CreatedAt.IsZero() {
		t.Fatal(
			"expected created_at to be populated",
		)
	}

	if created.UpdatedAt.IsZero() {
		t.Fatal(
			"expected updated_at to be populated",
		)
	}

	getResp := communityAPIRequest(
		t,
		app,
		http.MethodGet,
		communityIDPath(created.ID),
		"",
	)

	if getResp.StatusCode != http.StatusOK {
		defer getResp.Body.Close()

		t.Fatalf(
			"expected GET status %d, got %d",
			http.StatusOK,
			getResp.StatusCode,
		)
	}

	retrieved := decodeCommunity(
		t,
		getResp,
	)

	if retrieved.ID != created.ID {
		t.Fatalf(
			"expected retrieved ID %d, got %d",
			created.ID,
			retrieved.ID,
		)
	}

	if retrieved.Slug != created.Slug {
		t.Fatalf(
			"expected retrieved slug %q, got %q",
			created.Slug,
			retrieved.Slug,
		)
	}
}

// -----------------------------------------------------------------------------
// POST duplicate
// -----------------------------------------------------------------------------

func TestCommunityAPI_CreateDuplicate(t *testing.T) {
	app := newCommunityAPITestApp(t)

	body := `{
		"name": "Toronto Men",
		"slug": "duplicate-community-integration",
		"description": "Integration test community.",
		"external_source": "manual"
	}`

	firstResp := communityAPIRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	if firstResp.StatusCode != http.StatusCreated {
		defer firstResp.Body.Close()

		t.Fatalf(
			"expected first POST status %d, got %d",
			http.StatusCreated,
			firstResp.StatusCode,
		)
	}

	firstResp.Body.Close()

	secondResp := communityAPIRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	defer secondResp.Body.Close()

	if secondResp.StatusCode != http.StatusConflict {
		t.Fatalf(
			"expected duplicate POST status %d, got %d",
			http.StatusConflict,
			secondResp.StatusCode,
		)
	}
}

// -----------------------------------------------------------------------------
// GET existing
// -----------------------------------------------------------------------------

func TestCommunityAPI_GetExisting(t *testing.T) {
	app := newCommunityAPITestApp(t)

	body := `{
		"name": "Mississauga Men",
		"slug": "mississauga-men-integration",
		"description": "A community for men in Mississauga.",
		"external_source": "manual"
	}`

	createResp := communityAPIRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	if createResp.StatusCode != http.StatusCreated {
		defer createResp.Body.Close()

		t.Fatalf(
			"expected POST status %d, got %d",
			http.StatusCreated,
			createResp.StatusCode,
		)
	}

	created := decodeCommunity(
		t,
		createResp,
	)

	getResp := communityAPIRequest(
		t,
		app,
		http.MethodGet,
		communityIDPath(created.ID),
		"",
	)

	if getResp.StatusCode != http.StatusOK {
		defer getResp.Body.Close()

		t.Fatalf(
			"expected GET status %d, got %d",
			http.StatusOK,
			getResp.StatusCode,
		)
	}

	retrieved := decodeCommunity(
		t,
		getResp,
	)

	if retrieved.ID != created.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			created.ID,
			retrieved.ID,
		)
	}

	if retrieved.Name != "Mississauga Men" {
		t.Fatalf(
			"expected name %q, got %q",
			"Mississauga Men",
			retrieved.Name,
		)
	}
}

// -----------------------------------------------------------------------------
// GET missing
// -----------------------------------------------------------------------------

func TestCommunityAPI_GetMissing(t *testing.T) {
	app := newCommunityAPITestApp(t)

	resp := communityAPIRequest(
		t,
		app,
		http.MethodGet,
		"/api/v1/communities/999999999",
		"",
	)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected GET missing status %d, got %d",
			http.StatusNotFound,
			resp.StatusCode,
		)
	}
}

// -----------------------------------------------------------------------------
// PUT existing
// -----------------------------------------------------------------------------

func TestCommunityAPI_UpdateExisting(t *testing.T) {
	app := newCommunityAPITestApp(t)

	createBody := `{
		"name": "Hamilton Men",
		"slug": "hamilton-men-integration",
		"description": "Original description.",
		"external_source": "manual"
	}`

	createResp := communityAPIRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		createBody,
	)

	if createResp.StatusCode != http.StatusCreated {
		defer createResp.Body.Close()

		t.Fatalf(
			"expected POST status %d, got %d",
			http.StatusCreated,
			createResp.StatusCode,
		)
	}

	created := decodeCommunity(
		t,
		createResp,
	)

	updateBody := `{
		"name": "Hamilton Men's Community",
		"slug": "hamilton-men-integration",
		"description": "Updated description.",
		"external_source": "manual"
	}`

	updateResp := communityAPIRequest(
		t,
		app,
		http.MethodPut,
		communityIDPath(created.ID),
		updateBody,
	)

	if updateResp.StatusCode != http.StatusOK {
		defer updateResp.Body.Close()

		t.Fatalf(
			"expected PUT status %d, got %d",
			http.StatusOK,
			updateResp.StatusCode,
		)
	}

	updated := decodeCommunity(
		t,
		updateResp,
	)

	if updated.ID != created.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			created.ID,
			updated.ID,
		)
	}

	if updated.Name != "Hamilton Men's Community" {
		t.Fatalf(
			"expected updated name %q, got %q",
			"Hamilton Men's Community",
			updated.Name,
		)
	}

	if updated.Description != "Updated description." {
		t.Fatalf(
			"expected updated description %q, got %q",
			"Updated description.",
			updated.Description,
		)
	}
}

// -----------------------------------------------------------------------------
// DELETE existing
// -----------------------------------------------------------------------------

func TestCommunityAPI_DeleteExisting(t *testing.T) {
	app := newCommunityAPITestApp(t)

	createBody := `{
		"name": "Burlington Men",
		"slug": "burlington-men-integration",
		"description": "A community for men in Burlington.",
		"external_source": "manual"
	}`

	createResp := communityAPIRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		createBody,
	)

	if createResp.StatusCode != http.StatusCreated {
		defer createResp.Body.Close()

		t.Fatalf(
			"expected POST status %d, got %d",
			http.StatusCreated,
			createResp.StatusCode,
		)
	}

	created := decodeCommunity(
		t,
		createResp,
	)

	deleteResp := communityAPIRequest(
		t,
		app,
		http.MethodDelete,
		communityIDPath(created.ID),
		"",
	)

	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf(
			"expected DELETE status %d, got %d",
			http.StatusNoContent,
			deleteResp.StatusCode,
		)
	}

	getResp := communityAPIRequest(
		t,
		app,
		http.MethodGet,
		communityIDPath(created.ID),
		"",
	)

	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected GET after DELETE status %d, got %d",
			http.StatusNotFound,
			getResp.StatusCode,
		)
	}
}
