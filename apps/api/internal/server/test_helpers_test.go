package server

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/amyismebyme/the-village/apps/api/internal/handlers"
	"github.com/amyismebyme/the-village/apps/api/internal/health"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

type routerCommunityServiceMock struct{}

func (routerCommunityServiceMock) Create(
	_ context.Context,
	community *model.Community,
) error {
	community.ID = 1

	return nil
}

func (routerCommunityServiceMock) Get(
	_ context.Context,
	_ int64,
) (*model.Community, error) {
	return nil, nil
}

func (routerCommunityServiceMock) List(
	_ context.Context,
) ([]*model.Community, error) {
	return []*model.Community{}, nil
}

func (routerCommunityServiceMock) Update(
	_ context.Context,
	_ *model.Community,
) error {
	return nil
}

func (routerCommunityServiceMock) Delete(
	_ context.Context,
	_ int64,
) error {
	return nil
}

var _ service.CommunityService = routerCommunityServiceMock{}

// newTestRouter builds a full router with middleware for tests
// that need to verify middleware behavior (logging, request IDs).
func newTestRouter() http.Handler {
	logger := slog.New(
		slog.NewTextHandler(
			httptest.NewRecorder(),
			nil,
		),
	)

	healthRegistry := health.NewRegistry()

	handler := handlers.NewHandler(
		routerCommunityServiceMock{},
	)

	return NewRouter(
		logger,
		healthRegistry,
		handler,
	)
}

// newTestRouterWithLogs builds a router that captures structured
// log output into the supplied buffer.
func newTestRouterWithLogs(logs *bytes.Buffer) http.Handler {
	logger := slog.New(
		slog.NewTextHandler(
			logs,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		),
	)

	healthRegistry := health.NewRegistry()

	handler := handlers.NewHandler(
		routerCommunityServiceMock{},
	)

	return NewRouter(
		logger,
		healthRegistry,
		handler,
	)
}
