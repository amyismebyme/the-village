package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
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
		limit int,
		offset int,
	) (service.CommunityListResult, error)

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
	limit int,
	offset int,
) (service.CommunityListResult, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}

	return service.CommunityListResult{
		Communities: []*model.Community{},
		Limit:       limit,
		Offset:      offset,
	}, nil
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
// Helpers
// -----------------------------------------------------------------------------

func newCommunityRequest(
	method string,
	target string,
	body string,
) *http.Request {
	req := httptest.NewRequest(
		method,
		target,
		strings.NewReader(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	return req
}

func setCommunityID(
	req *http.Request,
	id string,
) {
	req.SetPathValue("id", id)
}

func decodeCommunityResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
) communityResponse {
	t.Helper()

	var response communityResponse

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode community response: %v",
			err,
		)
	}

	return response
}

func decodeCommunityListResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
) []communityResponse {
	t.Helper()

	var response struct {
		Communities []communityResponse `json:"communities"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"decode community list response: %v",
			err,
		)
	}

	return response.Communities
}

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

	req := newCommunityRequest(
		http.MethodPost,
		"/api/v1/communities",
		body,
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

	if received.ID != 1 {
		t.Fatalf(
			"expected service-populated ID 1, got %d",
			received.ID,
		)
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

	response := decodeCommunityResponse(
		t,
		recorder,
	)

	if response.ID != received.ID {
		t.Fatalf(
			"expected response ID %d, got %d",
			received.ID,
			response.ID,
		)
	}

	if response.Name != received.Name {
		t.Fatalf(
			"expected response name %q, got %q",
			received.Name,
			response.Name,
		)
	}

	if response.Slug != received.Slug {
		t.Fatalf(
			"expected response slug %q, got %q",
			received.Slug,
			response.Slug,
		)
	}
}

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

	req := newCommunityRequest(
		http.MethodPost,
		"/api/v1/communities",
		`{"name":`,
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

	req := newCommunityRequest(
		http.MethodPost,
		"/api/v1/communities",
		body,
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
		t.Fatal("expected service to be called")
	}
}

func TestCreateCommunityDuplicate(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return service.ErrCommunityAlreadyExists
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men"
	}`

	req := newCommunityRequest(
		http.MethodPost,
		"/api/v1/communities",
		body,
	)

	recorder := httptest.NewRecorder()

	handler.CreateCommunity(
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

	req := newCommunityRequest(
		http.MethodPost,
		"/api/v1/communities",
		body,
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

func TestCreateCommunityMultipleJSONValues(t *testing.T) {
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
		"slug": "toronto-men"
	} {
		"name": "Second Community"
	}`

	req := newCommunityRequest(
		http.MethodPost,
		"/api/v1/communities",
		body,
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
			"expected service not to be called for multiple JSON values",
		)
	}
}

func TestCreateCommunityMissingContentType(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/communities",
		strings.NewReader(body),
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
			limit int,
			offset int,
		) (service.CommunityListResult, error) {
			return service.CommunityListResult{
				Communities: expected,
				Limit:       limit,
				Offset:      offset,
				Total:       int64(len(expected)),
			}, nil
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

	communities := decodeCommunityListResponse(
		t,
		recorder,
	)

	if len(communities) != 2 {
		t.Fatalf(
			"expected 2 communities, got %d",
			len(communities),
		)
	}

	if communities[0].ID != 1 {
		t.Fatalf(
			"expected first ID 1, got %d",
			communities[0].ID,
		)
	}

	if communities[0].Slug != "toronto-men" {
		t.Fatalf(
			"expected first slug %q, got %q",
			"toronto-men",
			communities[0].Slug,
		)
	}

	if communities[1].ID != 2 {
		t.Fatalf(
			"expected second ID 2, got %d",
			communities[1].ID,
		)
	}

	if communities[1].Slug != "mississauga-men" {
		t.Fatalf(
			"expected second slug %q, got %q",
			"mississauga-men",
			communities[1].Slug,
		)
	}
}

func TestListCommunitiesEmpty(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) (service.CommunityListResult, error) {
			return service.CommunityListResult{
				Communities: []*model.Community{},
				Limit:       limit,
				Offset:      offset,
			}, nil
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

	communities := decodeCommunityListResponse(
		t,
		recorder,
	)

	if communities == nil {
		t.Fatal(
			"expected communities to be an empty array, got nil",
		)
	}

	if len(communities) != 0 {
		t.Fatalf(
			"expected 0 communities, got %d",
			len(communities),
		)
	}
}

func TestListCommunitiesNilEntriesAreSkipped(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) (service.CommunityListResult, error) {
			return service.CommunityListResult{
				Communities: []*model.Community{
					{
						ID:   1,
						Name: "Toronto Men",
						Slug: "toronto-men",
					},
					nil,
					{
						ID:   2,
						Name: "Mississauga Men",
						Slug: "mississauga-men",
					},
				},
				Limit:  limit,
				Offset: offset,
				Total:  2,
			}, nil
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

	communities := decodeCommunityListResponse(
		t,
		recorder,
	)

	if len(communities) != 2 {
		t.Fatalf(
			"expected 2 communities after skipping nil entry, got %d",
			len(communities),
		)
	}
}

func TestListCommunitiesServiceError(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) (service.CommunityListResult, error) {
			return service.CommunityListResult{}, errors.New("database unavailable")
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

	setCommunityID(req, "1")

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

	response := decodeCommunityResponse(
		t,
		recorder,
	)

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

	if response.Description != expected.Description {
		t.Fatalf(
			"expected description %q, got %q",
			expected.Description,
			response.Description,
		)
	}

	if response.ExternalSource != expected.ExternalSource {
		t.Fatalf(
			"expected external source %q, got %q",
			expected.ExternalSource,
			response.ExternalSource,
		)
	}
}

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

			setCommunityID(req, id)

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

	setCommunityID(req, "999")

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

func TestGetCommunityNilResult(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		getFunc: func(
			ctx context.Context,
			id int64,
		) (*model.Community, error) {
			return nil, nil
		},
	}

	handler := NewHandler(mockService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities/999",
		nil,
	)

	setCommunityID(req, "999")

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

// -----------------------------------------------------------------------------
// PUT /api/v1/communities/:id
// -----------------------------------------------------------------------------

func TestUpdateCommunity(t *testing.T) {
	t.Parallel()

	var received *model.Community

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			received = community

			// Simulate values populated by the service/repository.
			community.ID = 1

			return nil
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men's Community",
		"slug": "toronto-men",
		"description": "Updated description.",
		"external_source": "manual"
	}`

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		body,
	)

	setCommunityID(req, "1")

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

	if received == nil {
		t.Fatal(
			"expected community to be passed to service",
		)
	}

	if received.ID != 1 {
		t.Fatalf(
			"expected community ID 1, got %d",
			received.ID,
		)
	}

	if received.Name != "Toronto Men's Community" {
		t.Fatalf(
			"unexpected name %q",
			received.Name,
		)
	}

	if received.Slug != "toronto-men" {
		t.Fatalf(
			"unexpected slug %q",
			received.Slug,
		)
	}

	if received.Description != "Updated description." {
		t.Fatalf(
			"unexpected description %q",
			received.Description,
		)
	}

	if received.ExternalSource != "manual" {
		t.Fatalf(
			"unexpected external source %q",
			received.ExternalSource,
		)
	}

	response := decodeCommunityResponse(
		t,
		recorder,
	)

	if response.ID != received.ID {
		t.Fatalf(
			"expected response ID %d, got %d",
			received.ID,
			response.ID,
		)
	}

	if response.Name != received.Name {
		t.Fatalf(
			"expected response name %q, got %q",
			received.Name,
			response.Name,
		)
	}

	if response.Slug != received.Slug {
		t.Fatalf(
			"expected response slug %q, got %q",
			received.Slug,
			response.Slug,
		)
	}
}

func TestUpdateCommunityInvalidID(t *testing.T) {
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

			req := newCommunityRequest(
				http.MethodPut,
				"/api/v1/communities/"+id,
				body,
			)

			setCommunityID(req, id)

			recorder := httptest.NewRecorder()

			handler.UpdateCommunity(
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
					"service should not be called for invalid ID",
				)
			}
		})
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

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		`{"name":`,
	)

	setCommunityID(req, "1")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
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
			"service should not be called for malformed JSON",
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

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		body,
	)

	setCommunityID(req, "1")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
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
			"expected service not to be called for unknown JSON field",
		)
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

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		body,
	)

	setCommunityID(req, "1")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
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
			"expected service not to be called for multiple JSON values",
		)
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

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		body,
	)

	setCommunityID(req, "1")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
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

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		body,
	)

	setCommunityID(req, "1")

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

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/999",
		body,
	)

	setCommunityID(req, "999")

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

func TestUpdateCommunityDatabaseFailure(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		updateFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return errors.New("database unavailable")
		},
	}

	handler := NewHandler(mockService)

	body := `{
		"name": "Toronto Men's Community",
		"slug": "toronto-men"
	}`

	req := newCommunityRequest(
		http.MethodPut,
		"/api/v1/communities/1",
		body,
	)

	setCommunityID(req, "1")

	recorder := httptest.NewRecorder()

	handler.UpdateCommunity(
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
// DELETE /api/v1/communities/:id
// -----------------------------------------------------------------------------

func TestDeleteCommunity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var deletedID int64

		mockService := &communityServiceMock{
			deleteFunc: func(
				ctx context.Context,
				id int64,
			) error {
				deletedID = id
				return nil
			},
		}

		handler := NewHandler(mockService)

		req := httptest.NewRequest(
			http.MethodDelete,
			"/api/v1/communities/1",
			nil,
		)

		setCommunityID(req, "1")

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

		if deletedID != 1 {
			t.Fatalf(
				"expected delete ID 1, got %d",
				deletedID,
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
					deleteFunc: func(
						ctx context.Context,
						id int64,
					) error {
						serviceCalled = true
						return nil
					},
				}

				handler := NewHandler(mockService)

				req := httptest.NewRequest(
					http.MethodDelete,
					"/api/v1/communities/"+id,
					nil,
				)

				setCommunityID(req, id)

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

				if serviceCalled {
					t.Fatal(
						"expected service not to be called for invalid ID",
					)
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

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

		setCommunityID(req, "999")

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
		t.Parallel()

		mockService := &communityServiceMock{
			deleteFunc: func(
				ctx context.Context,
				id int64,
			) error {
				return errors.New("database unavailable")
			},
		}

		handler := NewHandler(mockService)

		req := httptest.NewRequest(
			http.MethodDelete,
			"/api/v1/communities/1",
			nil,
		)

		setCommunityID(req, "1")

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

			handler.CreateCommunity(
				recorder,
				req,
			)

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

			handler.ListCommunities(
				recorder,
				req,
			)

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

			setCommunityID(req, "1")

			recorder := httptest.NewRecorder()

			handler.GetCommunity(
				recorder,
				req,
			)

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

			setCommunityID(req, "1")

			recorder := httptest.NewRecorder()

			handler.UpdateCommunity(
				recorder,
				req,
			)

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

			setCommunityID(req, "1")

			recorder := httptest.NewRecorder()

			handler.DeleteCommunity(
				recorder,
				req,
			)

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

func TestCreateCommunityValidationContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing name",
			body: `{
				"name": "",
				"slug": "valid-community"
			}`,
		},
		{
			name: "invalid slug",
			body: `{
				"name": "Toronto Men",
				"slug": "Invalid Slug"
			}`,
		},
		{
			name: "name too short",
			body: `{
				"name": "ab",
				"slug": "valid-community"
			}`,
		},
		{
			name: "description too long",
			body: `{
				"name": "Toronto Men",
				"slug": "valid-community",
				"description": "` + strings.Repeat("x", 1001) + `"
			}`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := &communityServiceMock{
				createFunc: func(
					_ context.Context,
					_ *model.Community,
				) error {
					return service.ErrInvalidCommunity
				},
			}

			handler := NewHandler(mockService)

			req := newCommunityRequest(
				http.MethodPost,
				"/api/v1/communities",
				tt.body,
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

			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf(
					"expected application/json, got %q",
					got,
				)
			}

			body := recorder.Body.String()

			if !strings.Contains(
				body,
				`"code":"invalid_community"`,
			) {
				t.Fatalf(
					"expected invalid_community error, got %s",
					body,
				)
			}
		})
	}
}
