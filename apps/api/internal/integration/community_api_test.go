//go:build integration

package integration

import (
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	"net/http"
	"net/http/httptest"
	"testing"
)

// -----------------------------------------------------------------------------
// Test application
// -----------------------------------------------------------------------------

type communityAPITestApp struct {
	server *httptest.Server
	repo   *postgres.CommunityRepository
	db     interface {
		Close()
	}
}

func TestCommunityAPI_CreateThenGet(t *testing.T) {
	app := newIntegrationApp(t)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men-integration",
		"description": "A community for men in Toronto.",
		"external_source": "manual"
	}`

	createResp := integrationRequest(
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

	created := decodeCommunity(t, createResp)

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

	getResp := integrationRequest(
		t,
		app,
		http.MethodGet,
		communityPath(created.ID),
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

	retrieved := decodeCommunity(t, getResp)

	if retrieved.ID != created.ID {
		t.Fatalf(
			"expected retrieved ID %d, got %d",
			created.ID,
			retrieved.ID,
		)
	}

	if retrieved.Name != created.Name {
		t.Fatalf(
			"expected retrieved name %q, got %q",
			created.Name,
			retrieved.Name,
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
