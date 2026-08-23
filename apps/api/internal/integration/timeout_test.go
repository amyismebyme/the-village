//go:build integration

package integration

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
	"github.com/amyismebyme/the-village/apps/api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCommunityAPIDatabaseTimeoutReturns504(t *testing.T) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		),
	)

	app := newIntegrationAppWithLogger(t, logger)

	// Use the real PostgreSQL pool, but deliberately impose a very short
	// repository timeout around a pg_sleep query so the HTTP request observes
	// a database-originated context deadline.
	timeoutRepo := &testTimeoutCommunityRepository{
		CommunityRepository: app.repo,
		pool:                app.db.Pool(),
		timeout:             25 * time.Millisecond,
	}

	communityService := service.NewCommunityService(timeoutRepo)
	handler := handlers.NewHandler(
		communityService,
		logger,
	)

	testServer := httptest.NewServer(
		server.NewRouter(
			logger,
			health.NewRegistry(),
			handler,
		),
	)
	defer testServer.Close()

	requestBody := `{
		"name": "Database Timeout Test",
		"slug": "database-timeout-test",
		"description": "Should produce a 504",
		"external_source": "integration"
	}`

	req, err := http.NewRequest(
		http.MethodPost,
		testServer.URL+"/api/v1/communities",
		strings.NewReader(requestBody),
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := testServer.Client().Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/communities: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusGatewayTimeout {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusGatewayTimeout,
			resp.StatusCode,
			body,
		)
	}

	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		t.Fatal("expected X-Request-ID on timeout response")
	}

	output := logs.String()

	for _, expected := range []string{
		"operation=create",
		"status=504",
		"duration_ms=",
		"request_id=" + requestID,
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf(
				"expected timeout log to contain %q; got:\n%s",
				expected,
				output,
			)
		}
	}
}

type testTimeoutCommunityRepository struct {
	*postgres.CommunityRepository
	pool    *pgxpool.Pool
	timeout time.Duration
}

func (r *testTimeoutCommunityRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (*model.Community, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var value int

	err := r.pool.QueryRow(
		ctx,
		`SELECT 1 FROM pg_sleep(1)`,
	).Scan(&value)

	if err != nil {
		return nil, err
	}

	return nil, nil
}
