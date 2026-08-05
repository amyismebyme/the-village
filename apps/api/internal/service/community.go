package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
)

var (
	ErrCommunityAlreadyExists = errors.New("community already exists")
	ErrInvalidCommunity       = errors.New("invalid community")
	ErrInvalidCommunityID     = errors.New("invalid community id")
	ErrNilCommunity           = errors.New("community is required")
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
	communityRepository repository.CommunityRepository,
) CommunityService {
	if communityRepository == nil {
		panic("community service: repository is required")
	}

	return &communityService{
		repository: communityRepository,
	}
}

func (s *communityService) Create(
	ctx context.Context,
	community *model.Community,
) error {
	if community == nil {
		return ErrNilCommunity
	}

	normalizeCommunity(community)

	if err := community.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCommunity, err)
	}

	existing, err := s.repository.FindBySlug(ctx, community.Slug)

	switch {
	case err == nil && existing != nil:
		return fmt.Errorf(
			"%w: slug %q",
			ErrCommunityAlreadyExists,
			community.Slug,
		)

	case err != nil && !errors.Is(err, repository.ErrNotFound):
		return fmt.Errorf(
			"community service: check slug %q: %w",
			community.Slug,
			err,
		)
	}

	if err := s.repository.Create(ctx, community); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return fmt.Errorf(
				"%w: slug %q",
				ErrCommunityAlreadyExists,
				community.Slug,
			)
		}

		return fmt.Errorf(
			"community service: create community: %w",
			err,
		)
	}

	return nil
}

func (s *communityService) Get(
	ctx context.Context,
	id int64,
) (*model.Community, error) {
	if id <= 0 {
		return nil, ErrInvalidCommunityID
	}

	community, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(
			"community service: get community %d: %w",
			id,
			err,
		)
	}

	return community, nil
}

func (s *communityService) List(
	ctx context.Context,
) ([]*model.Community, error) {
	communities, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"community service: list communities: %w",
			err,
		)
	}

	if communities == nil {
		return []*model.Community{}, nil
	}

	return communities, nil
}

func (s *communityService) Update(
	ctx context.Context,
	community *model.Community,
) error {
	if community == nil {
		return ErrNilCommunity
	}

	if community.ID <= 0 {
		return ErrInvalidCommunityID
	}

	normalizeCommunity(community)

	if err := community.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCommunity, err)
	}

	existing, err := s.repository.FindBySlug(ctx, community.Slug)

	switch {
	case err == nil && existing != nil && existing.ID != community.ID:
		return fmt.Errorf(
			"%w: slug %q",
			ErrCommunityAlreadyExists,
			community.Slug,
		)

	case err != nil && !errors.Is(err, repository.ErrNotFound):
		return fmt.Errorf(
			"community service: check slug %q: %w",
			community.Slug,
			err,
		)
	}

	if err := s.repository.Update(ctx, community); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return fmt.Errorf(
				"%w: slug %q",
				ErrCommunityAlreadyExists,
				community.Slug,
			)
		}

		return fmt.Errorf(
			"community service: update community %d: %w",
			community.ID,
			err,
		)
	}

	return nil
}

func (s *communityService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrInvalidCommunityID
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf(
			"community service: delete community %d: %w",
			id,
			err,
		)
	}

	return nil
}

func normalizeCommunity(community *model.Community) {
	community.Name = strings.TrimSpace(community.Name)

	community.Slug = strings.ToLower(
		strings.TrimSpace(community.Slug),
	)

	community.Description = strings.TrimSpace(community.Description)
	community.ExternalSource = strings.TrimSpace(community.ExternalSource)
}
