package server

import (
	"context"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

type routerCommunityServiceMock struct{}

func (routerCommunityServiceMock) Create(
	_ context.Context,
	_ *model.Community,
) error {
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
