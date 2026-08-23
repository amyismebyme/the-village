package service

import (
	"context"
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type cancelAwareCommunityRepository struct {
	*mockCommunityRepository
}

func (r *cancelAwareCommunityRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (*model.Community, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return r.mockCommunityRepository.FindBySlug(ctx, slug)
}

func TestCommunityServicePreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &cancelAwareCommunityRepository{
		mockCommunityRepository: newMockCommunityRepository(),
	}

	svc := NewCommunityService(repo)

	err := svc.Create(
		ctx,
		&model.Community{
			Name: "Toronto Men",
			Slug: "toronto-men",
		},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected context.Canceled to survive service layer, got %v",
			err,
		)
	}
}
