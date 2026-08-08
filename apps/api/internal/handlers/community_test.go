package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

type communityServiceMock struct {
	createFunc func(
		ctx context.Context,
		community *model.Community,
	) error

	createdCommunity *model.Community
}

func (m *communityServiceMock) Create(
	ctx context.Context,
	community *model.Community,
) error {
	m.createdCommunity = community

	if m.createFunc != nil {
		return m.createFunc(ctx, community)
	}

	return nil
}

func (m *communityServiceMock) Get(
	context.Context,
	int64,
) (*model.Community, error) {
	return nil, nil
}

func (m *communityServiceMock) List(
	context.Context,
) ([]*model.Community, error) {
	return nil, nil
}

func (m *communityServiceMock) Update(
	context.Context,
	*model.Community,
) error {
	return nil
}

func (m *communityServiceMock) Delete(
	context.Context,
	int64,
) error {
	return nil
}

func TestCreateCommunity(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			community.ID = 1

			return nil
		},
	}

	handler := &Handler{
		communityService: mockService,
	}

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men",
		"description": "A community for men in Toronto.",
		"external_source": "manual"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/communities",
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

	if mockService.createdCommunity == nil {
		t.Fatal("expected community to be passed to service")
	}

	if mockService.createdCommunity.Name != "Toronto Men" {
		t.Fatalf(
			"expected name %q, got %q",
			"Toronto Men",
			mockService.createdCommunity.Name,
		)
	}

	if mockService.createdCommunity.Slug != "toronto-men" {
		t.Fatalf(
			"expected slug %q, got %q",
			"toronto-men",
			mockService.createdCommunity.Slug,
		)
	}
}

func TestCreateCommunityInvalidJSON(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{}

	handler := &Handler{
		communityService: mockService,
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/communities",
		strings.NewReader(`{"name":`),
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

	if mockService.createdCommunity != nil {
		t.Fatal("service should not be called for malformed JSON")
	}
}

func TestCreateCommunityUnknownField(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{}

	handler := &Handler{
		communityService: mockService,
	}

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men",
		"unknown": "should fail"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/communities",
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

func TestCreateCommunityDuplicateSlug(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return service.ErrCommunityAlreadyExists
		},
	}

	handler := &Handler{
		communityService: mockService,
	}

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/communities",
		strings.NewReader(body),
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

func TestCreateCommunityValidationError(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return service.ErrInvalidCommunity
		},
	}

	handler := &Handler{
		communityService: mockService,
	}

	body := `{
		"name": "x",
		"slug": "bad"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/communities",
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

func TestCreateCommunityServiceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")

	mockService := &communityServiceMock{
		createFunc: func(
			ctx context.Context,
			community *model.Community,
		) error {
			return expectedErr
		},
	}

	handler := &Handler{
		communityService: mockService,
	}

	body := `{
		"name": "Toronto Men",
		"slug": "toronto-men"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/communities",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.CreateCommunity(
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

func TestCreateCommunityMethodNotAllowed(t *testing.T) {
	t.Parallel()

	mockService := &communityServiceMock{}

	handler := &Handler{
		communityService: mockService,
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/communities",
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
}
