package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"strings"
)

const (
	DefaultCommunityPageLimit = 20
	MaxCommunityPageLimit     = 100
)

type CommunityListResult struct {
	Communities []*model.Community
	Limit       int
	Offset      int
	Total       int64
}

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
		limit int,
		offset int,
	) (CommunityListResult, error)

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

	community.Normalize()

	if err := validateCommunity(community); err != nil {
		return err
	}

	existing, err := s.slugExists(ctx, community.Slug)
	if err != nil {
		return fmt.Errorf(
			"community service: check slug %q: %w",
			community.Slug,
			err,
		)
	}

	if existing != nil {
		return fmt.Errorf(
			"%w: slug %q",
			ErrCommunityAlreadyExists,
			community.Slug,
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

	metrics.CommunityCreateTotal.
		WithLabelValues("success").
		Inc()

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
	limit int,
	offset int,
) (CommunityListResult, error) {
	if limit == 0 {
		limit = DefaultCommunityPageLimit
	}

	if limit < 0 || limit > MaxCommunityPageLimit {
		return CommunityListResult{}, ErrInvalidPagination
	}

	if offset < 0 {
		return CommunityListResult{}, ErrInvalidPagination
	}

	communities, total, err := s.repository.List(
		ctx,
		limit,
		offset,
	)
	if err != nil {
		return CommunityListResult{}, fmt.Errorf(
			"community service: list communities: %w",
			err,
		)
	}

	if communities == nil {
		communities = []*model.Community{}
	}

	return CommunityListResult{
		Communities: communities,
		Limit:       limit,
		Offset:      offset,
		Total:       total,
	}, nil
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

	community.Normalize()

	if err := validateCommunity(community); err != nil {
		return err
	}

	current, err := s.repository.FindByID(ctx, community.ID)
	if err != nil {
		return fmt.Errorf(
			"community service: find community %d: %w",
			community.ID,
			err,
		)
	}

	// Only check slug uniqueness if it has changed.
	if current.Slug != community.Slug {
		existing, err := s.slugExists(ctx, community.Slug)
		if err != nil {
			return fmt.Errorf(
				"community service: check slug %q: %w",
				community.Slug,
				err,
			)
		}

		if existing != nil && existing.ID != community.ID {
			return fmt.Errorf(
				"%w: slug %q",
				ErrCommunityAlreadyExists,
				community.Slug,
			)
		}
	}

	if err := s.repository.Update(
		ctx,
		community,
	); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}

		if errors.Is(err, repository.ErrAlreadyExists) {
			return fmt.Errorf(
				"%w: slug %q",
				ErrCommunityAlreadyExists,
				community.Slug,
			)
		}

		return fmt.Errorf(
			"community service: update community: %w",
			err,
		)
	}

	metrics.CommunityUpdateTotal.
		WithLabelValues("success").
		Inc()

	return nil

}

func (s *communityService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrInvalidCommunityID
	}

	if _, err := s.repository.FindByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}

		return fmt.Errorf(
			"community service: find community %d: %w",
			id,
			err,
		)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}

		return fmt.Errorf(
			"community service: delete community: %w",
			err,
		)
	}

	metrics.CommunityDeleteTotal.
		WithLabelValues("success").
		Inc()

	return nil
}

func validateCommunity(
	c *model.Community,
) error {
	if c == nil {
		return fmt.Errorf(
			"%w: community is nil",
			ErrInvalidCommunity,
		)
	}

	if err := c.Validate(); err != nil {
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues(validationField(err)).
			Inc()

		return fmt.Errorf(
			"%w: %w",
			ErrInvalidCommunity,
			err,
		)
	}

	return nil
}

func (s *communityService) slugExists(
	ctx context.Context,
	slug string,
) (*model.Community, error) {

	community, err := s.repository.FindBySlug(ctx, slug)

	switch {

	case err == nil:
		return community, nil

	case errors.Is(err, repository.ErrNotFound):
		return nil, nil

	default:
		return nil, err
	}
}

func validationField(err error) string {
	if err == nil {
		return "unknown"
	}

	message := err.Error()

	field, _, ok := strings.Cut(
		message,
		":",
	)

	if !ok {
		return "unknown"
	}

	switch field {
	case "name":
		return "name"

	case "slug":
		return "slug"

	case "description":
		return "description"

	case "external_source":
		return "external_source"

	default:
		return "unknown"
	}
}
