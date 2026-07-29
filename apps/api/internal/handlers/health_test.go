package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/testutil"
)

type mockChecker struct {
	name string
	err  error
}

func (m mockChecker) Name() string {
	return m.name
}

func (m mockChecker) Check(ctx context.Context) error {
	return m.err
}

func TestHealthHandler_Healthy(t *testing.T) {
	registry := health.NewRegistry()

	registry.Register(mockChecker{
		name: "test",
		err:  nil,
	})

	SetHealthRegistry(registry)

	req := testutil.NewRequest(http.MethodGet, "/health")
	rr := testutil.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
	}

	var response HealthResponse
	testutil.DecodeJSON(t, rr.Body.Bytes(), &response)

	if response.Status != "healthy" {
		t.Fatalf("expected healthy got %q", response.Status)
	}

	if response.Checks["test"] != "ok" {
		t.Fatalf("expected test check to be ok")
	}
}