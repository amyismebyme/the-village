//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestCommunityAPI_InvalidResources(t *testing.T) {
	t.Run("invalid ID", func(t *testing.T) {
		app := newIntegrationApp(t)

		resp := integrationRequest(
			t,
			app,
			http.MethodGet,
			"/api/v1/communities/abc",
			"",
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusBadRequest,
		)
		requireRequestID(t, resp)

		resp.Body.Close()
	})

	t.Run("missing GET resource", func(t *testing.T) {
		app := newIntegrationApp(t)

		resp := integrationRequest(
			t,
			app,
			http.MethodGet,
			"/api/v1/communities/999999999",
			"",
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusNotFound,
		)
		requireRequestID(t, resp)

		resp.Body.Close()
	})

	t.Run("missing PUT resource", func(t *testing.T) {
		app := newIntegrationApp(t)

		body := `{
			"name": "Missing Community",
			"slug": "missing-community",
			"description": "Should not exist"
		}`

		resp := integrationRequest(
			t,
			app,
			http.MethodPut,
			"/api/v1/communities/999999999",
			body,
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusNotFound,
		)
		requireRequestID(t, resp)

		resp.Body.Close()
	})

	t.Run("missing DELETE resource", func(t *testing.T) {
		app := newIntegrationApp(t)

		resp := integrationRequest(
			t,
			app,
			http.MethodDelete,
			"/api/v1/communities/999999999",
			"",
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusNotFound,
		)
		requireRequestID(t, resp)

		resp.Body.Close()
	})
}

func TestCommunityAPI_DeleteSemantics(t *testing.T) {
	app := newIntegrationApp(t)

	createBody := `{"name":"Delete Semantics","slug":"delete-semantics","description":"Delete semantics test"}`
	createResp := integrationRequest(t, app, http.MethodPost, "/api/v1/communities", createBody)
	requireJSONResponse(t, createResp, http.StatusCreated)
	created := decodeCommunity(t, createResp)

	firstDelete := integrationRequest(t, app, http.MethodDelete, communityPath(created.ID), "")
	if firstDelete.StatusCode != http.StatusNoContent {
		defer firstDelete.Body.Close()
		t.Fatalf("expected first DELETE status %d, got %d", http.StatusNoContent, firstDelete.StatusCode)
	}
	firstDelete.Body.Close()

	secondDelete := integrationRequest(t, app, http.MethodDelete, communityPath(created.ID), "")
	requireJSONResponse(t, secondDelete, http.StatusNotFound)
	requireRequestID(t, secondDelete)
	secondDelete.Body.Close()
}
