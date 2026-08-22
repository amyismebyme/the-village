package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

func TestListCommunitiesPaginationDefaults(t *testing.T) {
	var gotLimit, gotOffset int

	handler := NewHandler(&communityServiceMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) (service.CommunityListResult, error) {
			gotLimit, gotOffset = limit, offset
			return service.CommunityListResult{
				Communities: []*model.Community{},
				Limit:       limit,
				Offset:      offset,
				Total:       42,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/communities", nil)
	recorder := httptest.NewRecorder()

	handler.ListCommunities(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if gotLimit != service.DefaultCommunityPageLimit || gotOffset != 0 {
		t.Fatalf("expected limit=%d offset=0, got limit=%d offset=%d", service.DefaultCommunityPageLimit, gotLimit, gotOffset)
	}

	var body communityListResponse
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Pagination.Limit != service.DefaultCommunityPageLimit || body.Pagination.Offset != 0 || body.Pagination.Total != 42 {
		t.Fatalf("unexpected pagination: %+v", body.Pagination)
	}
}

func TestListCommunitiesPaginationQuery(t *testing.T) {
	var gotLimit, gotOffset int

	handler := NewHandler(&communityServiceMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) (service.CommunityListResult, error) {
			gotLimit, gotOffset = limit, offset
			return service.CommunityListResult{
				Communities: []*model.Community{},
				Limit:       limit,
				Offset:      offset,
				Total:       7,
			}, nil
		},
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/communities?limit=5&offset=10",
		nil,
	)
	recorder := httptest.NewRecorder()

	handler.ListCommunities(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if gotLimit != 5 || gotOffset != 10 {
		t.Fatalf("expected limit=5 offset=10, got limit=%d offset=%d", gotLimit, gotOffset)
	}
}

func TestListCommunitiesRejectsInvalidPagination(t *testing.T) {
	tests := []string{
		"?limit=abc",
		"?limit=0",
		"?limit=-1",
		"?offset=abc",
		"?offset=-1",
	}

	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			handler := NewHandler(&communityServiceMock{})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/communities"+query, nil)
			recorder := httptest.NewRecorder()

			handler.ListCommunities(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
		})
	}
}

func TestListCommunitiesRejectsExcessiveLimit(t *testing.T) {
	handler := NewHandler(&communityServiceMock{
		listFunc: func(
			ctx context.Context,
			limit int,
			offset int,
		) (service.CommunityListResult, error) {
			return service.CommunityListResult{}, service.ErrInvalidPagination
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/communities?limit=101", nil)
	recorder := httptest.NewRecorder()

	handler.ListCommunities(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
