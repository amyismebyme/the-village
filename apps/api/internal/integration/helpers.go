//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"log/slog"

	"github.com/amyismebyme/the-village/apps/api/internal/config"
	"github.com/amyismebyme/the-village/apps/api/internal/database"
	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/logger"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository/postgres"
	"github.com/amyismebyme/the-village/apps/api/internal/server"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
)

const (
	databaseWaitTimeout = 30 * time.Second
	databaseRetryDelay  = 1 * time.Second
)

type integrationCommunityResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description,omitempty"`
	ExternalSource string    `json:"external_source,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type integrationCommunityListResponse struct {
	Communities []integrationCommunityResponse `json:"communities"`
}

var integrationEnvOnce sync.Once

// -----------------------------------------------------------------------------
// Shared integration environment
// -----------------------------------------------------------------------------

// loadIntegrationEnv loads the integration environment once.
//
// .env.integration lives beside the integration tests under:
//
//	internal/integration/.env.integration
//
// godotenv.Overload is intentional here because the integration environment
// should take precedence over an existing local environment when running
// integration tests.
func loadIntegrationEnv() {
	integrationEnvOnce.Do(func() {
		if err := godotenv.Overload(".env.integration"); err != nil {
			fmt.Printf(
				"warning: unable to load .env.integration: %v\n",
				err,
			)
		}
	})
}

// MustGetEnv returns an environment variable or panics when it is missing.
func MustGetEnv(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf(
			"environment variable %s not set",
			key,
		))
	}

	return value
}

// -----------------------------------------------------------------------------
// Database harness
// -----------------------------------------------------------------------------

// OpenTestDatabase opens the real PostgreSQL database configured for
// integration testing and waits until PostgreSQL is reachable.
//
// The returned database is automatically closed when the test finishes.
func OpenTestDatabase(t *testing.T) *database.Database {
	t.Helper()

	loadIntegrationEnv()

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		databaseWaitTimeout,
	)
	defer cancel()

	db, err := database.Open(ctx, cfg.Database)
	if err != nil {
		t.Fatalf(
			"open integration database: %v",
			err,
		)
	}

	if err := WaitForDatabase(ctx, db); err != nil {
		db.Close()

		t.Fatalf(
			"database never became ready: %v",
			err,
		)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// WaitForDatabase waits until the application's database connection can
// successfully ping PostgreSQL.
func WaitForDatabase(
	ctx context.Context,
	db *database.Database,
) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	// Try immediately instead of unnecessarily waiting for the first ticker.
	if err := db.Ping(ctx); err == nil {
		return nil
	}

	timeout := time.NewTimer(databaseWaitTimeout)
	defer timeout.Stop()

	ticker := time.NewTicker(databaseRetryDelay)
	defer ticker.Stop()

	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf(
				"waiting for database: %w",
				ctx.Err(),
			)

		case <-timeout.C:
			if lastErr != nil {
				return fmt.Errorf(
					"timed out waiting for database: %w",
					lastErr,
				)
			}

			return fmt.Errorf(
				"timed out waiting for database",
			)

		case <-ticker.C:
			err := db.Ping(ctx)

			if err == nil {
				return nil
			}

			lastErr = err
		}
	}
}

// -----------------------------------------------------------------------------
// Shared application harness
// -----------------------------------------------------------------------------

// integrationApp represents the complete HTTP application used by integration
// tests:
//
//	HTTP
//	  ↓
//	Handler
//	  ↓
//	Service
//	  ↓
//	Repository
//	  ↓
//	PostgreSQL
type integrationApp struct {
	server *httptest.Server
	db     *database.Database
	repo   *postgres.CommunityRepository
}

// newIntegrationApp creates a complete application backed by the real
// PostgreSQL database.

func newIntegrationApp(t *testing.T) *integrationApp {
	t.Helper()

	cfg := config.Load()

	return newIntegrationAppWithLogger(
		t,
		logger.New(cfg),
	)
}

// -----------------------------------------------------------------------------
// HTTP helpers
// -----------------------------------------------------------------------------

func integrationRequest(
	t *testing.T,
	app *integrationApp,
	method string,
	path string,
	body string,
) *http.Response {
	t.Helper()

	req, err := http.NewRequest(
		method,
		app.server.URL+path,
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf(
			"create %s request: %v",
			method,
			err,
		)
	}

	if body != "" {
		req.Header.Set(
			"Content-Type",
			"application/json",
		)
	}

	resp, err := app.server.Client().Do(req)
	if err != nil {
		t.Fatalf(
			"execute %s %s: %v",
			method,
			path,
			err,
		)
	}

	return resp
}

// communityAPIRequest is kept as a small compatibility wrapper so existing
// Community API tests do not need to duplicate HTTP setup.
func communityAPIRequest(
	t *testing.T,
	app *integrationApp,
	method string,
	path string,
	body string,
) *http.Response {
	t.Helper()

	return integrationRequest(
		t,
		app,
		method,
		path,
		body,
	)
}

// -----------------------------------------------------------------------------
// JSON response helpers
// -----------------------------------------------------------------------------

func decodeCommunity(
	t *testing.T,
	resp *http.Response,
) integrationCommunityResponse {
	t.Helper()

	defer resp.Body.Close()

	var community integrationCommunityResponse

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&community); err != nil {
		t.Fatalf(
			"decode community response: %v",
			err,
		)
	}

	return community
}

type communityListResponse struct {
	Communities []*model.Community `json:"communities"`
}

func decodeCommunityList(
	t *testing.T,
	resp *http.Response,
) integrationCommunityListResponse {
	t.Helper()

	defer resp.Body.Close()

	var community integrationCommunityListResponse

	if err := json.NewDecoder(resp.Body).Decode(&community); err != nil {
		t.Fatalf(
			"decode community list response: %v",
			err,
		)
	}

	return community
}

// -----------------------------------------------------------------------------
// URL helpers
// -----------------------------------------------------------------------------

func communityPath(id int64) string {
	return "/api/v1/communities/" +
		strconv.FormatInt(id, 10)
}

// Backward-compatible name for existing tests.
func communityIDPath(id int64) string {
	return communityPath(id)
}

func newIntegrationAppWithLogger(
	t *testing.T,
	appLogger *slog.Logger,
) *integrationApp {
	t.Helper()

	loadIntegrationEnv()

	db := OpenTestDatabase(t)

	metrics.Register(
		prometheus.DefaultRegisterer,
		db.Pool(),
	)

	repo := postgres.NewCommunityRepository(
		db.Pool(),
	)

	ctx := context.Background()

	if err := repo.DeleteAll(ctx); err != nil {
		t.Fatalf(
			"clean communities before test: %v",
			err,
		)
	}

	t.Cleanup(func() {
		if err := repo.DeleteAll(context.Background()); err != nil {
			t.Logf(
				"clean communities after test: %v",
				err,
			)
		}
	})

	communityService := service.NewCommunityService(repo)

	handler := handlers.NewHandler(
		communityService,
		appLogger,
	)

	healthRegistry := health.NewRegistry()

	httpHandler := server.NewRouter(
		appLogger,
		healthRegistry,
		handler,
	)

	testServer := httptest.NewServer(httpHandler)

	t.Cleanup(func() {
		testServer.Close()
	})

	return &integrationApp{
		server: testServer,
		db:     db,
		repo:   repo,
	}
}

func requireJSONResponse(
	t *testing.T,
	resp *http.Response,
	expectedStatus int,
) {
	t.Helper()

	if resp.StatusCode != expectedStatus {
		defer resp.Body.Close()

		t.Fatalf(
			"expected HTTP status %d, got %d",
			expectedStatus,
			resp.StatusCode,
		)
	}

	contentType := resp.Header.Get("Content-Type")

	if contentType != "application/json" {
		defer resp.Body.Close()

		t.Fatalf(
			"expected Content-Type application/json, got %q",
			contentType,
		)
	}
}

func requireRequestID(
	t *testing.T,
	resp *http.Response,
) string {
	t.Helper()

	requestID := resp.Header.Get("X-Request-ID")

	if requestID == "" {
		defer resp.Body.Close()

		t.Fatal("expected X-Request-ID header")
	}

	return requestID
}
