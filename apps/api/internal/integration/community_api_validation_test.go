//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)


func TestCommunityAPI_ValidationFailures(t *testing.T) {
	t.Run("missing required name", func(t *testing.T) {
		app := newIntegrationApp(t)

		before, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities before request: %v",
				err,
			)
		}

		resp := integrationRequest(
			t,
			app,
			http.MethodPost,
			"/api/v1/communities",
			`{
				"name": "",
				"slug": "valid-community"
			}`,
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusBadRequest,
		)
		requireRequestID(t, resp)

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			t.Fatalf(
				"read validation response: %v",
				err,
			)
		}

		if !strings.Contains(
			string(body),
			`"code":"invalid_community"`,
		) {
			t.Fatalf(
				"expected invalid_community error, got %s",
				body,
			)
		}

		after, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities after request: %v",
				err,
			)
		}

		if len(after) != len(before) {
			t.Fatalf(
				"validation failure changed database state: before=%d after=%d",
				len(before),
				len(after),
			)
		}
	})

	t.Run("invalid slug", func(t *testing.T) {
		app := newIntegrationApp(t)

		before, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities before request: %v",
				err,
			)
		}

		resp := integrationRequest(
			t,
			app,
			http.MethodPost,
			"/api/v1/communities",
			`{
				"name": "Toronto Men",
				"slug": "Invalid Slug"
			}`,
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusBadRequest,
		)
		requireRequestID(t, resp)

		resp.Body.Close()

		after, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities after request: %v",
				err,
			)
		}

		if len(after) != len(before) {
			t.Fatalf(
				"invalid slug changed database state: before=%d after=%d",
				len(before),
				len(after),
			)
		}
	})

	t.Run("name too short", func(t *testing.T) {
		app := newIntegrationApp(t)

		before, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities before request: %v",
				err,
			)
		}

		resp := integrationRequest(
			t,
			app,
			http.MethodPost,
			"/api/v1/communities",
			`{
				"name": "ab",
				"slug": "valid-community"
			}`,
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusBadRequest,
		)
		requireRequestID(t, resp)

		resp.Body.Close()

		after, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities after request: %v",
				err,
			)
		}

		if len(after) != len(before) {
			t.Fatalf(
				"short name changed database state: before=%d after=%d",
				len(before),
				len(after),
			)
		}
	})

	t.Run("name too long", func(t *testing.T) {
		app := newIntegrationApp(t)

		before, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities before request: %v",
				err,
			)
		}

		body := fmt.Sprintf(
			`{
				"name": "%s",
				"slug": "valid-community"
			}`,
			strings.Repeat("a", 101),
		)

		resp := integrationRequest(
			t,
			app,
			http.MethodPost,
			"/api/v1/communities",
			body,
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusBadRequest,
		)
		requireRequestID(t, resp)

		resp.Body.Close()

		after, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities after request: %v",
				err,
			)
		}

		if len(after) != len(before) {
			t.Fatalf(
				"long name changed database state: before=%d after=%d",
				len(before),
				len(after),
			)
		}
	})

	t.Run("description too long", func(t *testing.T) {
		app := newIntegrationApp(t)

		before, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities before request: %v",
				err,
			)
		}

		body := fmt.Sprintf(
			`{
				"name": "Toronto Men",
				"slug": "valid-community",
				"description": "%s"
			}`,
			strings.Repeat("d", 2001),
		)

		resp := integrationRequest(
			t,
			app,
			http.MethodPost,
			"/api/v1/communities",
			body,
		)

		requireJSONResponse(
			t,
			resp,
			http.StatusBadRequest,
		)
		requireRequestID(t, resp)

		resp.Body.Close()

		after, err := app.repo.List(context.Background())
		if err != nil {
			t.Fatalf(
				"list communities after request: %v",
				err,
			)
		}

		if len(after) != len(before) {
			t.Fatalf(
				"long description changed database state: before=%d after=%d",
				len(before),
				len(after),
			)
		}
	})
}



func TestCommunityAPI_DuplicateSlug(t *testing.T) {
	app := newIntegrationApp(t)

	firstBody := `{
		"name": "Duplicate Slug One",
		"slug": "duplicate-slug-integration",
		"description": "First community",
		"external_source": "integration"
	}`

	firstResp := integrationRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		firstBody,
	)

	requireJSONResponse(
		t,
		firstResp,
		http.StatusCreated,
	)
	requireRequestID(t, firstResp)

	first := decodeCommunity(t, firstResp)

	before, err := app.repo.List(context.Background())
	if err != nil {
		t.Fatalf(
			"list communities before duplicate request: %v",
			err,
		)
	}

	secondBody := `{
		"name": "Duplicate Slug Two",
		"slug": "duplicate-slug-integration",
		"description": "Second community",
		"external_source": "integration"
	}`

	secondResp := integrationRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/communities",
		secondBody,
	)

	requireJSONResponse(
		t,
		secondResp,
		http.StatusConflict,
	)
	requireRequestID(t, secondResp)

	body, err := io.ReadAll(secondResp.Body)
	secondResp.Body.Close()

	if err != nil {
		t.Fatalf(
			"read duplicate response: %v",
			err,
		)
	}

	if !strings.Contains(
		string(body),
		`"code":"community_already_exists"`,
	) {
		t.Fatalf(
			"expected community_already_exists error, got %s",
			body,
		)
	}

	after, err := app.repo.List(context.Background())
	if err != nil {
		t.Fatalf(
			"list communities after duplicate request: %v",
			err,
		)
	}

	if len(after) != len(before) {
		t.Fatalf(
			"duplicate request changed database state: before=%d after=%d",
			len(before),
			len(after),
		)
	}

	stored, err := app.repo.FindByID(
		context.Background(),
		first.ID,
	)
	if err != nil {
		t.Fatalf(
			"verify original community after duplicate request: %v",
			err,
		)
	}

	if stored.Slug != first.Slug {
		t.Fatalf(
			"original community slug changed: expected %q got %q",
			first.Slug,
			stored.Slug,
		)
	}
}