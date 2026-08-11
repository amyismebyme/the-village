package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// Mock Community Service
// -----------------------------------------------------------------------------

type communityServiceMock struct {
	createFunc func(
		ctx context.Context,
		community *model.Community,
	) error

	getFunc func(
		ctx context.Context,
		id int64,
	) (*model.Community, error)

	listFunc func(
		ctx context.Context,
	) ([]*model.Community, error)

	updateFunc func(
		ctx context.Context,
		community *model.Community,
	) error

	deleteFunc func(
		ctx context.Context,
		id int64,
	) error
}

func (m *communityServiceMock) Create(
	ctx context.Context,
	community *model.Community,
) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, community)
	}

	return nil
}

func (m *communityServiceMock) Get(
	ctx context.Context,
	id int64,
) (*model.Community, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}

	return nil, nil
}

func (m *communityServiceMock) List(
	ctx context.Context,
) ([]*model.Community, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}

	return nil, nil
}

func (m *communityServiceMock) Update(
	ctx context.Context,
	community *model.Community,
) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, community)
	}

	return nil
}

func (m *communityServiceMock) Delete(
	ctx context.Context,
	id int64,
) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}

	return nil
}

var _ service.CommunityService = (*communityServiceMock)(nil)

// -----------------------------------------------------------------------------
// Handler constructor
// -----------------------------------------------------------------------------

func TestNewHandler(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{}

	handler := NewHandler(mockService)

	if handler == nil {
		t.Fatal("expected handler, got nil")
	}

	if handler.communityService == nil {
		t.Fatal("expected community service to be configured")
	}
}

// -----------------------------------------------------------------------------
// POST /api/v1/communities
// -----------------------------------------------------------------------------

func TestCreateCommunity(t *testing.T) {
	t.Parallel()

	var received *model.Community

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			received = community
			community.ID = 1

			return nil
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men",
		"description": "A community for men in Toronto.",
		"external_source": "manual"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/communities",
		strings.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
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

	if received == nil {
		t.Fatal("expected community to be passed to service")
	}

	if received.Name != "Toronto Men" {
		t.Fatalf(
			"expected name %q, got %q",
			"Toronto Men",
			received.Name,
		)
	}

	if received.Slug != "toronto-men" {
		t.Fatalf(
			"expected slug %q, got %q",
			"toronto-men",
			received.Slug,
		)
	}
}

// -----------------------------------------------------------------------------
// POST malformed JSON
// -----------------------------------------------------------------------------

func TestCreateCommunityMalformedJSON(t *testing.T) {
	t.Parallel()

	serviceCalled := false

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			serviceCalled = true
			return nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/communities",
		strings.NewReader(`{"name":`),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if serviceCalled {
		t.Fatal(
			"expected service not to be called for malformed JSON",
		)
	}
}

// -----------------------------------------------------------------------------
// POST invalid community
// -----------------------------------------------------------------------------

func TestCreateCommunityValidationError(t *testing.T) {
	t.Parallel()

	serviceCalled := false

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			serviceCalled = true

			if community.Name != "" {
				t.Errorf(
					"expected empty name, got %q",
					community.Name,
				)
			}

			if community.Slug != "" {
				t.Errorf(
					"expected empty slug, got %q",
					community.Slug,
				)
			}

			return service.ErrInvalidCommunity
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "",
		"slug": "",
		"description": ""
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/communities",
		strings.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if !serviceCalled {
		t.Fatal(
			"expected service to be called",
		)
	}
}

// -----------------------------------------------------------------------------
// GET /api/v1/communities
// -----------------------------------------------------------------------------

func TestListCommunities(t *testing.T) {
	t.Parallel()

	expected := []*model.Community{
		{
			ID:             1,
			Name:           "Toronto Men",
			Slug:           "toronto-men",
			Description:    "A community for men in Toronto.",
			ExternalSource: "manual",
		},
		{
			ID:             2,
			Name:           "Mississauga Men",
			Slug:           "mississauga-men",
			Description:    "A community for men in Mississauga.",
			ExternalSource: "manual",
		},
	}

	mockService := &communityServiceMock{
		listFunc: func(
			ctx context.Context,
		) ([]*model.Community, error) {
			return expected, nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListCommunities(
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

	var response struct {
		Communities []*model.Community `json:"communities"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if len(response.Communities) != 2 {
		t.Fatalf(
			"expected 2 communities, got %d",
			len(response.Communities),
		)
	}

	if response.Communities[0].Slug != "toronto-men" {
		t.Fatalf(
			"expected first slug %q, got %q",
			"toronto-men",
			response.Communities[0].Slug,
		)
	}

	if response.Communities[1].Slug != "mississauga-men" {
		t.Fatalf(
			"expected second slug %q, got %q",
			"mississauga-men",
			response.Communities[1].Slug,
		)
	}
}

// -----------------------------------------------------------------------------
// GET /api/v1/communities - empty result
// -----------------------------------------------------------------------------

func TestListCommunitiesEmpty(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		listFunc: func(
			ctx context.Context,
		) ([]*model.Community, error) {
			return nil, nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListCommunities(
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

	var response struct {
		Communities []*model.Community `json:"communities"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.Communities == nil {
		t.Fatal(
			"expected communities to be an empty array, got nil",
		)
	}

	if len(response.Communities) != 0 {
		t.Fatalf(
			"expected 0 communities, got %d",
			len(response.Communities),
		)
	}
}

// -----------------------------------------------------------------------------
// GET /api/v1/communities - service failure
// -----------------------------------------------------------------------------

func TestListCommunitiesServiceError(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		listFunc: func(
			ctx context.Context,
		) ([]*model.Community, error) {
			return nil, errors.New("database unavailable")
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ListCommunities(
		recorder,
		req,
	)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}
}

// -----------------------------------------------------------------------------
// POST /api/v1/communities - unknown field
// -----------------------------------------------------------------------------

func TestCreateCommunityUnknownField(t *testing.T) {
	t.Parallel()

	serviceCalled := false

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			serviceCalled = true
			return nil
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men",
		"description": "A community for men in Toronto.",
		"unexpected": "field"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/communities",
		strings.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.CreateCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if serviceCalled {
		t.Fatal(
			"expected service not to be called for unknown field",
		)
	}
}

// -----------------------------------------------------------------------------
// GET /api/v1/communities/:id
// -----------------------------------------------------------------------------

func TestGetCommunity(t *testing.T) {
	t.Parallel()

	expected := &model.Community{
		ID:             1,
		Name:           "Toronto Men",
		Slug:           "toronto-men",
		Description:    "A community for men in Toronto.",
		ExternalSource: "manual",
	}

	var requestedID int64

	mockService := &communityServiceMock{
		getFunc: func(
			ctx context.Context,
			id int64,
		) (*model.Community, error) {
			requestedID = id
			return expected, nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/1",
		nil,
	)

	req.SetPathValue("id", "1")

	recorder := httptest.NewRecorder()

	handler.GetCommunity(
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

	if requestedID != 1 {
		t.Fatalf(
			"expected service to receive ID 1, got %d",
			requestedID,
		)
	}

	var response model.Community

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.ID != expected.ID {
		t.Fatalf(
			"expected ID %d, got %d",
			expected.ID,
			response.ID,
		)
	}

	if response.Name != expected.Name {
		t.Fatalf(
			"expected name %q, got %q",
			expected.Name,
			response.Name,
		)
	}

	if response.Slug != expected.Slug {
		t.Fatalf(
			"expected slug %q, got %q",
			expected.Slug,
			response.Slug,
		)
	}
}

// -----------------------------------------------------------------------------
// GET /api/v1/communities/:id - invalid ID
// -----------------------------------------------------------------------------

func TestGetCommunityInvalidID(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"abc",
		"0",
		"-1",
	}

	for _, id := range testCases {
		id := id

		t.Run(id, func(t *testing.T) {
			t.Parallel()

			serviceCalled := false

			mockService := &communityServiceMock{
				getFunc: func(
					ctx context.Context,
					id int64,
				) (*model.Community, error) {
					serviceCalled = true
					return nil, nil
				},
			}

			handler := NewHandler(mockService)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/communities/"+id,
				nil,
			)

			req.SetPathValue("id", id)

			recorder := httptest.NewRecorder()

			handler.GetCommunity(
				recorder,
				req,
			)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusBadRequest,
					recorder.Code,
				)
			}

			if serviceCalled {
				t.Fatal(
					"expected service not to be called for invalid ID",
				)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// GET /api/v1/communities/:id - not found
// -----------------------------------------------------------------------------

func TestGetCommunityNotFound(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		getFunc: func(
			ctx context.Context,
			id int64,
		) (*model.Community, error) {
			return nil, repository.ErrNotFound
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/999",
		nil,
	)

	req.SetPathValue("id", "999")

	recorder := httptest.NewRecorder()

	handler.GetCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}

func TestUpdateCommunityDuplicate(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return service.ErrCommunityAlreadyExists
		},
	}

	handler := NewHandler(mockService)

	body := `{
    "name": "Toronto Men's Community",
    "slug": "toronto-men",
    "description": "Updated description.",
    "external_source": "manual"
}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		strings.NewReader(body),
	)

	req.SetPathValue("id", "1")

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusConflict,
			recorder.Code,
		)
	}
}

func TestUpdateCommunity(t *testing.T) {
	t.Parallel()

	updatedCommunity := &model.Community{
		ID:             1,
		Name:           "Toronto Men's Community",
		Slug:           "toronto-men",
		Description:    "Updated description.",
		ExternalSource: "manual",
	}

	var received *model.Community
	var getCalledWith int64

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			received = community
			return nil
		},
		getFunc: func(
			ctx context.Context,
			id int64,
		) (*model.Community, error) {
			getCalledWith = id
			return updatedCommunity, nil
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men's Community",
		"slug": "toronto-men",
		"description": "Updated description.",
		"external_source": "manual"
	}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		strings.NewReader(body),
	)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if received == nil {
		t.Fatal("expected community to be passed to service")
	}

	if received.ID != 1 {
		t.Fatalf("expected community ID 1, got %d", received.ID)
	}

	if received.Name != "Toronto Men's Community" {
		t.Fatalf("unexpected name %q", received.Name)
	}

	if received.Slug != "toronto-men" {
		t.Fatalf("unexpected slug %q", received.Slug)
	}

	if getCalledWith != 1 {
		t.Fatalf("expected Get to be called with ID 1, got %d", getCalledWith)
	}

	var response model.Community

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != updatedCommunity.ID {
		t.Fatalf(
			"expected response ID %d, got %d",
			updatedCommunity.ID,
			response.ID,
		)
	}

	if response.Name != updatedCommunity.Name {
		t.Fatalf(
			"expected response name %q, got %q",
			updatedCommunity.Name,
			response.Name,
		)
	}
}

func TestUpdateCommunityInvalidID(t *testing.T) {
	t.Parallel()

	serviceCalled := false

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			serviceCalled = true
			return nil
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men"
	}`

	testCases := []string{
		"abc",
		"0",
		"-1",
	}

	for _, id := range testCases {
		id := id

		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPut,
				"/api/v1/communities/"+id,
				strings.NewReader(body),
			)
			req.SetPathValue("id", id)

			recorder := httptest.NewRecorder()

			handler.UpdateCommunity(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusBadRequest,
					recorder.Code,
				)
			}
		})
	}

	if serviceCalled {
		t.Fatal("service should not be called for invalid IDs")
	}
}

func TestUpdateCommunityMalformedJSON(t *testing.T) {
	t.Parallel()

	serviceCalled := false

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			serviceCalled = true
			return nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		strings.NewReader(`{"name":`),
	)
	req.SetPathValue("id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if serviceCalled {
		t.Fatal("service should not be called for malformed JSON")
	}
}

func TestUpdateCommunityValidationFailure(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return service.ErrInvalidCommunity
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "",
		"slug": ""
	}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		strings.NewReader(body),
	)
	req.SetPathValue("id", "1")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestUpdateCommunityNotFound(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return repository.ErrNotFound
		},
	}

	handler := NewHandler(mockService)

	body := `{
    "name": "Toronto Men's Community",
    "slug": "toronto-men",
    "description": "Updated description.",
    "external_source": "manual"
}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/communities/999",
		strings.NewReader(body),
	)

	req.SetPathValue("id", "999")

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
		recorder,
		req,
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}

}

func TestDeleteCommunity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockService := &communityServiceMock{}

		handler := NewHandler(mockService)

		req := httptest.NewRequest(
			http.MethodDelete,
			"/api/v1/communities/1",
			nil,
		)

		req.SetPathValue("id", "1")

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

		if recorder.Body.Len() != 0 {
			t.Fatalf(
				"expected empty response body, got %q",
				recorder.Body.String(),
			)
		}
	})

	t.Run("invalid ID", func(t *testing.T) {
		testCases := []string{
			"abc",
			"0",
			"-1",
		}

		for _, id := range testCases {
			t.Run(id, func(t *testing.T) {
				mockService := &communityServiceMock{}

				handler := NewHandler(mockService)

				req := httptest.NewRequest(
					http.MethodDelete,
					"/api/v1/communities/"+id,
					nil,
				)

				req.SetPathValue("id", id)

				recorder := httptest.NewRecorder()

				handler.DeleteCommunity(
					recorder,
					req,
				)

				if recorder.Code != http.StatusBadRequest {
					t.Fatalf(
						"expected status %d, got %d",
						http.StatusBadRequest,
						recorder.Code,
					)
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockService := &communityServiceMock{
			deleteFunc: func(
				ctx context.Context,
				id int64,
			) error {
				return fmt.Errorf(
					"community service: delete community %d: %w",
					id,
					repository.ErrNotFound,
				)
			},
		}

		handler := NewHandler(mockService)

		req := httptest.NewRequest(
			http.MethodDelete,
			"/api/v1/communities/999",
			nil,
		)

		req.SetPathValue("id", "999")

		recorder := httptest.NewRecorder()

		handler.DeleteCommunity(
			recorder,
			req,
		)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusNotFound,
				recorder.Code,
			)
		}
	})

	t.Run("database failure", func(t *testing.T) {
		mockService := &communityServiceMock{
			deleteFunc: func(ctx context.Context, id int64) error {
				//return repository.ErrNotFound enable once everyting  is up
				return errors.New("database unavailable")

			},
		}

		handler := NewHandler(mockService)

		req := httptest.NewRequest(
			http.MethodDelete,
			"/api/v1/communities/1",
			nil,
		)

		req.SetPathValue("id", "1")

		recorder := httptest.NewRecorder()

		handler.DeleteCommunity(
			recorder,
			req,
		)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				recorder.Code,
			)
		}
	})
}

func TestCreateCommunityMissingContentType(t *testing.T) {
	mockService := &communityServiceMock{}

	handler := NewHandler(mockService)

	body := strings.NewReader(`{
		"name": "Toronto Men",
		"slug": "toronto-men",
		"description": "A community for men in Toronto.",
		"external_source": "manual"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	rec := httptest.NewRecorder()

	handler.CreateCommunity(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestUpdateCommunityUnknownField(t *testing.T) {
	t.Parallel()

	serviceCalled := false

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			serviceCalled = true
			return nil
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men",
		"banana": "hello"
	}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		strings.NewReader(body),
	)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if serviceCalled {
		t.Fatal("expected service not to be called for unknown JSON field")
	}
}

func TestUpdateCommunityMultipleJSONValues(t *testing.T) {
	t.Parallel()

	serviceCalled := false

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			serviceCalled = true
			return nil
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men"
	} {
		"name": "Second Community"
	}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		strings.NewReader(body),
	)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if serviceCalled {
		t.Fatal("expected service not to be called for multiple JSON values")
	}
}

// -----------------------------------------------------------------------------
// Method enforcement
// -----------------------------------------------------------------------------

func TestCreateCommunityMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&communityServiceMock{})

	methods := []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		method := method

		t.Run(method, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				method,
				"/api/v1/communities",
				nil,
			)

			recorder := httptest.NewRecorder()

			handler.CreateCommunity(recorder, req)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusMethodNotAllowed,
					recorder.Code,
				)
			}

			if got := recorder.Header().Get("Allow"); got != http.MethodPost {
				t.Fatalf(
					"expected Allow header %q, got %q",
					http.MethodPost,
					got,
				)
			}
		})
	}
}

func TestListCommunitiesMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&communityServiceMock{})

	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		method := method

		t.Run(method, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				method,
				"/api/v1/communities",
				nil,
			)

			recorder := httptest.NewRecorder()

			handler.ListCommunities(recorder, req)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusMethodNotAllowed,
					recorder.Code,
				)
			}

			if got := recorder.Header().Get("Allow"); got != http.MethodGet {
				t.Fatalf(
					"expected Allow header %q, got %q",
					http.MethodGet,
					got,
				)
			}
		})
	}
}

func TestGetCommunityMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&communityServiceMock{})

	methods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		method := method

		t.Run(method, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				method,
				"/api/v1/communities/1",
				nil,
			)

			req.SetPathValue("id", "1")

			recorder := httptest.NewRecorder()

			handler.GetCommunity(recorder, req)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusMethodNotAllowed,
					recorder.Code,
				)
			}

			if got := recorder.Header().Get("Allow"); got != http.MethodGet {
				t.Fatalf(
					"expected Allow header %q, got %q",
					http.MethodGet,
					got,
				)
			}
		})
	}
}

func TestUpdateCommunityMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&communityServiceMock{})

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
	}

	for _, method := range methods {
		method := method

		t.Run(method, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				method,
				"/api/v1/communities/1",
				nil,
			)

			req.SetPathValue("id", "1")

			recorder := httptest.NewRecorder()

			handler.UpdateCommunity(recorder, req)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusMethodNotAllowed,
					recorder.Code,
				)
			}

			if got := recorder.Header().Get("Allow"); got != http.MethodPut {
				t.Fatalf(
					"expected Allow header %q, got %q",
					http.MethodPut,
					got,
				)
			}
		})
	}
}

func TestDeleteCommunityMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&communityServiceMock{})

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
	}

	for _, method := range methods {
		method := method

		t.Run(method, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(
				method,
				"/api/v1/communities/1",
				nil,
			)

			req.SetPathValue("id", "1")

			recorder := httptest.NewRecorder()

			handler.DeleteCommunity(recorder, req)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusMethodNotAllowed,
					recorder.Code,
				)
			}

			if got := recorder.Header().Get("Allow"); got != http.MethodDelete {
				t.Fatalf(
					"expected Allow header %q, got %q",
					http.MethodDelete,
					got,
				)
			}
		})
	}
}
