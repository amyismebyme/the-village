package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/jackc/pgx/v5"
	"strings"
)

var (
	ErrCommunityAlreadyExists = errors.New("community already exists")
)

type CommunityService interface {
	Create(
		ctx context.Context,
		community *model.Community,
	) error

	Get(
		ctx context.Context,
		id int64,
	) (*model.Community, error)

	List(
		ctx context.Context,
	) ([]*model.Community, error)

	Update(
		ctx context.Context,
		community *model.Community,
	) error

	Delete(
		ctx context.Context,
		id int64,
	) error
}

type communityService struct {
	repository repository.CommunityRepository
}

func NewCommunityService(
	repository repository.CommunityRepository,
) CommunityService {

	return &communityService{
		repository: repository,
	}
}

func (s *communityService) Create(
	ctx context.Context,
	community *model.Community,
) error {

	community.Slug = strings.ToLower(strings.TrimSpace(community.Slug))
	community.Name = strings.TrimSpace(community.Name)
	community.Description = strings.TrimSpace(community.Description)

	if err := community.Validate(); err != nil {
		return err
	}

	existing, err := s.repository.FindBySlug(ctx, community.Slug)

	if err == nil && existing != nil {
		return ErrCommunityAlreadyExists
	}

	if err != nil &&
		!errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check existing community: %w", err)
	}

	if err := s.repository.Create(ctx, community); err != nil {
		return fmt.Errorf("create community: %w", err)
	}

	return nil
}

func (s *communityService) Get(
	ctx context.Context,
	id int64,
) (*model.Community, error) {

	community, err := s.repository.FindByID(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("get community: %w", err)
	}

	return community, nil
}

func (s *communityService) List(
	ctx context.Context,
) ([]*model.Community, error) {

	communities, err := s.repository.List(ctx)

	if err != nil {
		return nil, fmt.Errorf("list communities: %w", err)
	}

	return communities, nil
}

func (s *communityService) Update(
	ctx context.Context,
	community *model.Community,
) error {

	community.Slug = strings.ToLower(strings.TrimSpace(community.Slug))

	if err := community.Validate(); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, community); err != nil {
		return fmt.Errorf("update community: %w", err)
	}

	return nil
}

func (s *communityService) Delete(
	ctx context.Context,
	id int64,
) error {

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete community: %w", err)
	}

	return nil
}
