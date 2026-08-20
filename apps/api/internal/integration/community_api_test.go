//go:build integration

package integration

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"net/http"
	"testing"
)

func TestCommunityAPI_CRUDLifecycle(t *testing.T) {
	app := newIntegrationApp(t)

	// -------------------------------------------------------------------------
	// POST
	// -------------------------------------------------------------------------

	createBody := `{
		"name": "  Toronto Men  ",
		"slug": "  Toronto-Men  ",
		"description": "  A community for men in Toronto.  ",
		"external_source": "  manual  "
	}`

	createResp := integrationRequest(
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

	if contentType := createResp.Header.Get("Content-Type"); contentType != "application/json" {
		defer createResp.Body.Close()

		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	if requestID := createResp.Header.Get("X-Request-ID"); requestID == "" {
		defer createResp.Body.Close()

		t.Fatal("expected X-Request-ID header")
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
			"expected normalized name %q, got %q",
			"Toronto Men",
			created.Name,
		)
	}

	if created.Slug != "toronto-men" {
		t.Fatalf(
			"expected normalized slug %q, got %q",
			"toronto-men",
			created.Slug,
		)
	}

	if created.Description != "A community for men in Toronto." {
		t.Fatalf(
			"expected normalized description %q, got %q",
			"A community for men in Toronto.",
			created.Description,
		)
	}

	if created.ExternalSource != "manual" {
		t.Fatalf(
			"expected normalized external_source %q, got %q",
			"manual",
			created.ExternalSource,
		)
	}

	if created.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be populated")
	}

	if created.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be populated")
	}

	// -------------------------------------------------------------------------
	// GET after POST
	// -------------------------------------------------------------------------

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

	if contentType := getResp.Header.Get("Content-Type"); contentType != "application/json" {
		defer getResp.Body.Close()

		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	if requestID := getResp.Header.Get("X-Request-ID"); requestID == "" {
		defer getResp.Body.Close()

		t.Fatal("expected X-Request-ID header on GET")
	}

	retrieved := decodeCommunity(
		t,
		getResp,
	)

	// Proves GET returned the persisted record rather than the POST response.
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

	if retrieved.Description != created.Description {
		t.Fatalf(
			"expected retrieved description %q, got %q",
			created.Description,
			retrieved.Description,
		)
	}

	if retrieved.ExternalSource != created.ExternalSource {
		t.Fatalf(
			"expected retrieved external_source %q, got %q",
			created.ExternalSource,
			retrieved.ExternalSource,
		)
	}

	if retrieved.CreatedAt.IsZero() {
		t.Fatal("expected GET created_at to be populated")
	}

	if retrieved.UpdatedAt.IsZero() {
		t.Fatal("expected GET updated_at to be populated")
	}

	// -------------------------------------------------------------------------
	// PUT
	// -------------------------------------------------------------------------

	updateBody := `{
		"name": "Toronto Men Updated",
		"slug": "toronto-men-updated",
		"description": "Updated Toronto community description.",
		"external_source": "manual-updated"
	}`

	updateResp := integrationRequest(
		t,
		app,
		http.MethodPut,
		communityPath(created.ID),
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

	if contentType := updateResp.Header.Get("Content-Type"); contentType != "application/json" {
		defer updateResp.Body.Close()

		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	if requestID := updateResp.Header.Get("X-Request-ID"); requestID == "" {
		defer updateResp.Body.Close()

		t.Fatal("expected X-Request-ID header on PUT")
	}

	updated := decodeCommunity(
		t,
		updateResp,
	)

	if updated.ID != created.ID {
		t.Fatalf(
			"expected updated ID %d, got %d",
			created.ID,
			updated.ID,
		)
	}

	if updated.Name != "Toronto Men Updated" {
		t.Fatalf(
			"expected updated name %q, got %q",
			"Toronto Men Updated",
			updated.Name,
		)
	}

	if updated.Slug != "toronto-men-updated" {
		t.Fatalf(
			"expected updated slug %q, got %q",
			"toronto-men-updated",
			updated.Slug,
		)
	}

	if updated.Description != "Updated Toronto community description." {
		t.Fatalf(
			"expected updated description %q, got %q",
			"Updated Toronto community description.",
			updated.Description,
		)
	}

	if updated.ExternalSource != "manual-updated" {
		t.Fatalf(
			"expected updated external_source %q, got %q",
			"manual-updated",
			updated.ExternalSource,
		)
	}

	if updated.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be populated")
	}

	// -------------------------------------------------------------------------
	// GET after PUT
	// -------------------------------------------------------------------------

	getUpdatedResp := integrationRequest(
		t,
		app,
		http.MethodGet,
		communityPath(created.ID),
		"",
	)

	if getUpdatedResp.StatusCode != http.StatusOK {
		defer getUpdatedResp.Body.Close()

		t.Fatalf(
			"expected GET after PUT status %d, got %d",
			http.StatusOK,
			getUpdatedResp.StatusCode,
		)
	}

	retrievedUpdated := decodeCommunity(
		t,
		getUpdatedResp,
	)

	// Proves the UPDATE was persisted to PostgreSQL.
	if retrievedUpdated.ID != updated.ID {
		t.Fatalf(
			"expected persisted ID %d, got %d",
			updated.ID,
			retrievedUpdated.ID,
		)
	}

	if retrievedUpdated.Name != updated.Name {
		t.Fatalf(
			"expected persisted name %q, got %q",
			updated.Name,
			retrievedUpdated.Name,
		)
	}

	if retrievedUpdated.Slug != updated.Slug {
		t.Fatalf(
			"expected persisted slug %q, got %q",
			updated.Slug,
			retrievedUpdated.Slug,
		)
	}

	if retrievedUpdated.Description != updated.Description {
		t.Fatalf(
			"expected persisted description %q, got %q",
			updated.Description,
			retrievedUpdated.Description,
		)
	}

	if retrievedUpdated.ExternalSource != updated.ExternalSource {
		t.Fatalf(
			"expected persisted external_source %q, got %q",
			updated.ExternalSource,
			retrievedUpdated.ExternalSource,
		)
	}

	if retrievedUpdated.CreatedAt.IsZero() {
		t.Fatal("expected persisted created_at to be populated")
	}

	if retrievedUpdated.UpdatedAt.IsZero() {
		t.Fatal("expected persisted updated_at to be populated")
	}

	// -------------------------------------------------------------------------
	// LIST
	// -------------------------------------------------------------------------

	listResp := integrationRequest(
		t,
		app,
		http.MethodGet,
		"/api/v1/communities",
		"",
	)

	if listResp.StatusCode != http.StatusOK {
		defer listResp.Body.Close()

		t.Fatalf(
			"expected LIST status %d, got %d",
			http.StatusOK,
			listResp.StatusCode,
		)
	}

	if contentType := listResp.Header.Get("Content-Type"); contentType != "application/json" {
		defer listResp.Body.Close()

		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	list := decodeCommunityList(
		t,
		listResp,
	)

	if list.Communities == nil {
		t.Fatal("expected communities array, got nil")
	}

	found := false

	for _, community := range list.Communities {
		if community.ID != created.ID {
			continue
		}

		found = true

		if community.Name != updated.Name {
			t.Fatalf(
				"expected listed name %q, got %q",
				updated.Name,
				community.Name,
			)
		}

		if community.Slug != updated.Slug {
			t.Fatalf(
				"expected listed slug %q, got %q",
				updated.Slug,
				community.Slug,
			)
		}

		if community.Description != updated.Description {
			t.Fatalf(
				"expected listed description %q, got %q",
				updated.Description,
				community.Description,
			)
		}

		if community.ExternalSource != updated.ExternalSource {
			t.Fatalf(
				"expected listed external_source %q, got %q",
				updated.ExternalSource,
				community.ExternalSource,
			)
		}

		break
	}

	if !found {
		t.Fatalf(
			"expected community ID %d to be present in LIST response",
			created.ID,
		)
	}

	// -------------------------------------------------------------------------
	// DELETE
	// -------------------------------------------------------------------------

	deleteResp := integrationRequest(
		t,
		app,
		http.MethodDelete,
		communityPath(created.ID),
		"",
	)

	if deleteResp.StatusCode != http.StatusNoContent {
		defer deleteResp.Body.Close()

		t.Fatalf(
			"expected DELETE status %d, got %d",
			http.StatusNoContent,
			deleteResp.StatusCode,
		)
	}

	deleteResp.Body.Close()

	// -------------------------------------------------------------------------
	// GET after DELETE → 404
	// -------------------------------------------------------------------------

	getDeletedResp := integrationRequest(
		t,
		app,
		http.MethodGet,
		communityPath(created.ID),
		"",
	)

	if getDeletedResp.StatusCode != http.StatusNotFound {
		defer getDeletedResp.Body.Close()

		t.Fatalf(
			"expected GET after DELETE status %d, got %d",
			http.StatusNotFound,
			getDeletedResp.StatusCode,
		)
	}

	getDeletedResp.Body.Close()
}

func TestCommunityAPI_ListEmpty(t *testing.T) {
	app := newIntegrationApp(t)

	resp := integrationRequest(
		t,
		app,
		http.MethodGet,
		"/api/v1/communities",
		"",
	)

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		t.Fatalf(
			"expected LIST status %d, got %d",
			http.StatusOK,
			resp.StatusCode,
		)
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		defer resp.Body.Close()

		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}

	list := decodeCommunityList(
		t,
		resp,
	)

	if list.Communities == nil {
		t.Fatal(
			"expected communities to be an empty array, got nil",
		)
	}

	if len(list.Communities) != 0 {
		t.Fatalf(
			"expected empty community list, got %d records",
			len(list.Communities),
		)
	}
}

func TestCommunityAPI_DatabaseStateVerification(t *testing.T) {
	app := newIntegrationApp(t)

	// -------------------------------------------------------------------------
	// POST through HTTP
	// -------------------------------------------------------------------------

	createBody := `{
		"name": "Database Verification Community",
		"slug": "database-verification-community",
		"description": "Created through the API for database verification.",
		"external_source": "integration"
	}`

	createResp := integrationRequest(
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

	if created.ID <= 0 {
		t.Fatalf(
			"expected created community ID > 0, got %d",
			created.ID,
		)
	}

	// -------------------------------------------------------------------------
	// Verify POST persisted the record
	// -------------------------------------------------------------------------

	stored, err := app.repo.FindByID(
		context.Background(),
		created.ID,
	)
	if err != nil {
		t.Fatalf(
			"FindByID after POST failed: %v",
			err,
		)
	}

	if stored == nil {
		t.Fatal(
			"expected database record after POST, got nil",
		)
	}

	if stored.ID != created.ID {
		t.Fatalf(
			"expected database ID %d, got %d",
			created.ID,
			stored.ID,
		)
	}

	if stored.Name != created.Name {
		t.Fatalf(
			"expected database name %q, got %q",
			created.Name,
			stored.Name,
		)
	}

	if stored.Slug != created.Slug {
		t.Fatalf(
			"expected database slug %q, got %q",
			created.Slug,
			stored.Slug,
		)
	}

	if stored.Description != created.Description {
		t.Fatalf(
			"expected database description %q, got %q",
			created.Description,
			stored.Description,
		)
	}

	if stored.ExternalSource != created.ExternalSource {
		t.Fatalf(
			"expected database external_source %q, got %q",
			created.ExternalSource,
			stored.ExternalSource,
		)
	}

	if stored.CreatedAt.IsZero() {
		t.Fatal(
			"expected database created_at to be populated",
		)
	}

	if stored.UpdatedAt.IsZero() {
		t.Fatal(
			"expected database updated_at to be populated",
		)
	}

	// -------------------------------------------------------------------------
	// PUT through HTTP
	// -------------------------------------------------------------------------

	updateBody := `{
		"name": "Database Verification Community Updated",
		"slug": "database-verification-community-updated",
		"description": "Updated through the API.",
		"external_source": "integration-updated"
	}`

	updateResp := integrationRequest(
		t,
		app,
		http.MethodPut,
		communityPath(created.ID),
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
			"expected updated ID %d, got %d",
			created.ID,
			updated.ID,
		)
	}

	// -------------------------------------------------------------------------
	// Verify PUT persisted the changes
	// -------------------------------------------------------------------------

	storedUpdated, err := app.repo.FindByID(
		context.Background(),
		created.ID,
	)
	if err != nil {
		t.Fatalf(
			"FindByID after PUT failed: %v",
			err,
		)
	}

	if storedUpdated == nil {
		t.Fatal(
			"expected database record after PUT, got nil",
		)
	}

	if storedUpdated.ID != created.ID {
		t.Fatalf(
			"expected database ID %d after PUT, got %d",
			created.ID,
			storedUpdated.ID,
		)
	}

	if storedUpdated.Name != updated.Name {
		t.Fatalf(
			"expected database name %q after PUT, got %q",
			updated.Name,
			storedUpdated.Name,
		)
	}

	if storedUpdated.Slug != updated.Slug {
		t.Fatalf(
			"expected database slug %q after PUT, got %q",
			updated.Slug,
			storedUpdated.Slug,
		)
	}

	if storedUpdated.Description != updated.Description {
		t.Fatalf(
			"expected database description %q after PUT, got %q",
			updated.Description,
			storedUpdated.Description,
		)
	}

	if storedUpdated.ExternalSource != updated.ExternalSource {
		t.Fatalf(
			"expected database external_source %q after PUT, got %q",
			updated.ExternalSource,
			storedUpdated.ExternalSource,
		)
	}

	if storedUpdated.UpdatedAt.IsZero() {
		t.Fatal(
			"expected database updated_at to be populated after PUT",
		)
	}

	// -------------------------------------------------------------------------
	// DELETE through HTTP
	// -------------------------------------------------------------------------

	deleteResp := integrationRequest(
		t,
		app,
		http.MethodDelete,
		communityPath(created.ID),
		"",
	)

	if deleteResp.StatusCode != http.StatusNoContent {
		defer deleteResp.Body.Close()

		t.Fatalf(
			"expected DELETE status %d, got %d",
			http.StatusNoContent,
			deleteResp.StatusCode,
		)
	}

	deleteResp.Body.Close()

	// -------------------------------------------------------------------------
	// Verify DELETE removed the database record
	// -------------------------------------------------------------------------

	_, err = app.repo.FindByID(
		context.Background(),
		created.ID,
	)

	if !errors.Is(
		err,
		repository.ErrNotFound,
	) {
		t.Fatalf(
			"expected repository.ErrNotFound after DELETE, got %v",
			err,
		)
	}
}
