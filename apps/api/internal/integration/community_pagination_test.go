//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestCommunityPagination(t *testing.T) {
	app := newIntegrationApp(t)

	communities := []struct {
		name string
		slug string
	}{
		{name: "Beta", slug: "pagination-beta"},
		{name: "Alpha", slug: "pagination-alpha"},
		{name: "Alpha", slug: "pagination-alpha-two"},
	}

	for _, want := range communities {
		resp := integrationRequest(
			t,
			app,
			http.MethodPost,
			"/api/v1/communities",
			`{"name":"`+want.name+`","slug":"`+want.slug+`"}`,
		)

		requireJSONResponse(t, resp, http.StatusCreated)
		_ = decodeCommunity(t, resp)
	}

	t.Run("default", func(t *testing.T) {
		resp := integrationRequest(t, app, http.MethodGet, "/api/v1/communities", "")
		requireJSONResponse(t, resp, http.StatusOK)
		body := decodeCommunityList(t, resp)

		if body.Pagination.Limit != 20 || body.Pagination.Offset != 0 || body.Pagination.Total != 3 {
			t.Fatalf("unexpected default pagination: %+v", body.Pagination)
		}
	})

	t.Run("custom page", func(t *testing.T) {
		resp := integrationRequest(t, app, http.MethodGet, "/api/v1/communities?limit=1&offset=1", "")
		requireJSONResponse(t, resp, http.StatusOK)
		body := decodeCommunityList(t, resp)

		if body.Pagination.Limit != 1 || body.Pagination.Offset != 1 || body.Pagination.Total != 3 {
			t.Fatalf("unexpected custom pagination: %+v", body.Pagination)
		}

		if len(body.Communities) != 1 {
			t.Fatalf("expected 1 community, got %d", len(body.Communities))
		}

		if body.Communities[0].Name != "Alpha" {
			t.Fatalf("expected first page item name Alpha, got %q", body.Communities[0].Name)
		}
	})

	t.Run("deterministic duplicate-name ordering", func(t *testing.T) {
		resp := integrationRequest(t, app, http.MethodGet, "/api/v1/communities?limit=3&offset=0", "")
		requireJSONResponse(t, resp, http.StatusOK)
		body := decodeCommunityList(t, resp)

		if len(body.Communities) != 3 {
			t.Fatalf("expected 3 communities, got %d", len(body.Communities))
		}

		if body.Communities[0].Name != "Alpha" || body.Communities[1].Name != "Alpha" || body.Communities[2].Name != "Beta" {
			t.Fatalf("unexpected ordering: %+v", body.Communities)
		}

		if body.Communities[0].ID >= body.Communities[1].ID {
			t.Fatalf("expected duplicate-name rows ordered by ID: %d before %d", body.Communities[0].ID, body.Communities[1].ID)
		}
	})

	t.Run("offset beyond total", func(t *testing.T) {
		resp := integrationRequest(t, app, http.MethodGet, "/api/v1/communities?limit=20&offset=100", "")
		requireJSONResponse(t, resp, http.StatusOK)
		body := decodeCommunityList(t, resp)

		if body.Pagination.Total != 3 {
			t.Fatalf("expected total 3, got %d", body.Pagination.Total)
		}

		if len(body.Communities) != 0 {
			t.Fatalf("expected empty page, got %d communities", len(body.Communities))
		}
	})
}

func TestCommunityPaginationRejectsInvalidValues(t *testing.T) {
	app := newIntegrationApp(t)

	queries := []string{
		"?limit=abc",
		"?limit=0",
		"?limit=-1",
		"?limit=101",
		"?offset=abc",
		"?offset=-1",
	}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			resp := integrationRequest(t, app, http.MethodGet, "/api/v1/communities"+query, "")
			requireJSONResponse(t, resp, http.StatusBadRequest)
			resp.Body.Close()
		})
	}
}
