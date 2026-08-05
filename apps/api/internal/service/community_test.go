package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
)

type mockCommunityRepository struct {
	items map[int64]*model.Community

	nextID int64

	createErr     error
	findByIDErr   error
	findBySlugErr error
	listErr       error
	updateErr     error
	deleteErr     error

	createCalled bool
	updateCalled bool
	deleteCalled bool
}

func newMockCommunityRepository() *mockCommunityRepository {
	return &mockCommunityRepository{
		items:  make(map[int64]*model.Community),
		nextID: 1,
	}
}

func (m *mockCommunityRepository) Create(
	_ context.Context,
	community *model.Community,
) error {
	m.createCalled = true

	if m.createErr != nil {
		return m.createErr
	}

	if community.ID == 0 {
		community.ID = m.nextID
		m.nextID++
	}

	m.items[community.ID] = cloneCommunity(community)

	return nil
}

func (m *mockCommunityRepository) FindByID(
	_ context.Context,
	id int64,
) (*model.Community, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}

	community, exists := m.items[id]
	if !exists {
		return nil, repository.ErrNotFound
	}

	return cloneCommunity(community), nil
}

func (m *mockCommunityRepository) FindBySlug(
	_ context.Context,
	slug string,
) (*model.Community, error) {
	if m.findBySlugErr != nil {
		return nil, m.findBySlugErr
	}

	for _, community := range m.items {
		if community.Slug == slug {
			return cloneCommunity(community), nil
		}
	}

	return nil, repository.ErrNotFound
}

func (m *mockCommunityRepository) List(
	_ context.Context,
) ([]*model.Community, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	communities := make(
		[]*model.Community,
		0,
		len(m.items),
	)

	for _, community := range m.items {
		communities = append(
			communities,
			cloneCommunity(community),
		)
	}

	return communities, nil
}

func (m *mockCommunityRepository) Update(
	_ context.Context,
	community *model.Community,
) error {
	m.updateCalled = true

	if m.updateErr != nil {
		return m.updateErr
	}

	if _, exists := m.items[community.ID]; !exists {
		return repository.ErrNotFound
	}

	m.items[community.ID] = cloneCommunity(community)

	return nil
}

func (m *mockCommunityRepository) Delete(
	_ context.Context,
	id int64,
) error {
	m.deleteCalled = true

	if m.deleteErr != nil {
		return m.deleteErr
	}

	if _, exists := m.items[id]; !exists {
		return repository.ErrNotFound
	}

	delete(m.items, id)

	return nil
}

func (m *mockCommunityRepository) DeleteAll(
	_ context.Context,
) error {
	clear(m.items)

	return nil
}

func cloneCommunity(
	community *model.Community,
) *model.Community {
	if community == nil {
		return nil
	}

	cloned := *community

	return &cloned
}

func validCommunity() *model.Community {
	return &model.Community{
		Name:           "Toronto Men",
		Slug:           "toronto-men",
		Description:    "Helping men build meaningful friendships.",
		ExternalSource: "internal",
	}
}

func TestCommunityServiceCreate(t *testing.T) {
	t.Run("creates a valid community", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		community := &model.Community{
			Name:           "  Toronto Men  ",
			Slug:           "  TORONTO-MEN  ",
			Description:    "  Helping men build friendships.  ",
			ExternalSource: "  internal  ",
		}

		err := service.Create(
			context.Background(),
			community,
		)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if !repo.createCalled {
			t.Fatal("expected repository Create to be called")
		}

		if community.ID == 0 {
			t.Fatal("expected repository to assign an ID")
		}

		if community.Name != "Toronto Men" {
			t.Fatalf(
				"expected normalized name, got %q",
				community.Name,
			)
		}

		if community.Slug != "toronto-men" {
			t.Fatalf(
				"expected normalized slug, got %q",
				community.Slug,
			)
		}

		if community.Description != "Helping men build friendships." {
			t.Fatalf(
				"expected normalized description, got %q",
				community.Description,
			)
		}

		if community.ExternalSource != "internal" {
			t.Fatalf(
				"expected normalized external source, got %q",
				community.ExternalSource,
			)
		}
	})

	t.Run("rejects nil community", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		err := service.Create(
			context.Background(),
			nil,
		)

		if !errors.Is(err, ErrNilCommunity) {
			t.Fatalf(
				"expected ErrNilCommunity, got %v",
				err,
			)
		}

		if repo.createCalled {
			t.Fatal("repository Create should not be called")
		}
	})

	t.Run("rejects invalid community", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		community := &model.Community{
			Name: "ab",
			Slug: "valid-slug",
		}

		err := service.Create(
			context.Background(),
			community,
		)

		if !errors.Is(err, ErrInvalidCommunity) {
			t.Fatalf(
				"expected ErrInvalidCommunity, got %v",
				err,
			)
		}

		if repo.createCalled {
			t.Fatal("repository Create should not be called")
		}
	})

	t.Run("rejects duplicate slug", func(t *testing.T) {
		repo := newMockCommunityRepository()

		repo.items[1] = &model.Community{
			ID:   1,
			Name: "Existing Community",
			Slug: "toronto-men",
		}

		service := NewCommunityService(repo)

		community := validCommunity()

		err := service.Create(
			context.Background(),
			community,
		)

		if !errors.Is(err, ErrCommunityAlreadyExists) {
			t.Fatalf(
				"expected ErrCommunityAlreadyExists, got %v",
				err,
			)
		}

		if repo.createCalled {
			t.Fatal("repository Create should not be called")
		}
	})

	t.Run("wraps slug lookup failure", func(t *testing.T) {
		repo := newMockCommunityRepository()
		repo.findBySlugErr = errors.New("database unavailable")

		service := NewCommunityService(repo)

		err := service.Create(
			context.Background(),
			validCommunity(),
		)

		if err == nil {
			t.Fatal("expected an error")
		}

		if !strings.Contains(err.Error(), "check slug") {
			t.Fatalf(
				"expected wrapped slug lookup error, got %v",
				err,
			)
		}

		if repo.createCalled {
			t.Fatal("repository Create should not be called")
		}
	})

	t.Run("maps repository duplicate error", func(t *testing.T) {
		repo := newMockCommunityRepository()
		repo.createErr = repository.ErrAlreadyExists

		service := NewCommunityService(repo)

		err := service.Create(
			context.Background(),
			validCommunity(),
		)

		if !errors.Is(err, ErrCommunityAlreadyExists) {
			t.Fatalf(
				"expected ErrCommunityAlreadyExists, got %v",
				err,
			)
		}
	})
}

func TestCommunityServiceGet(t *testing.T) {
	t.Run("gets an existing community", func(t *testing.T) {
		repo := newMockCommunityRepository()

		repo.items[10] = &model.Community{
			ID:   10,
			Name: "Toronto Men",
			Slug: "toronto-men",
		}

		service := NewCommunityService(repo)

		community, err := service.Get(
			context.Background(),
			10,
		)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if community.ID != 10 {
			t.Fatalf(
				"expected ID 10, got %d",
				community.ID,
			)
		}

		if community.Name != "Toronto Men" {
			t.Fatalf(
				"unexpected community name %q",
				community.Name,
			)
		}
	})

	t.Run("rejects invalid ID", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		community, err := service.Get(
			context.Background(),
			0,
		)

		if community != nil {
			t.Fatal("expected nil community")
		}

		if !errors.Is(err, ErrInvalidCommunityID) {
			t.Fatalf(
				"expected ErrInvalidCommunityID, got %v",
				err,
			)
		}
	})

	t.Run("preserves repository not found error", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		_, err := service.Get(
			context.Background(),
			999,
		)

		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf(
				"expected repository.ErrNotFound, got %v",
				err,
			)
		}
	})
}

func TestCommunityServiceList(t *testing.T) {
	t.Run("lists communities", func(t *testing.T) {
		repo := newMockCommunityRepository()

		repo.items[1] = &model.Community{
			ID:   1,
			Name: "Toronto Men",
			Slug: "toronto-men",
		}

		repo.items[2] = &model.Community{
			ID:   2,
			Name: "Hamilton Men",
			Slug: "hamilton-men",
		}

		service := NewCommunityService(repo)

		communities, err := service.List(
			context.Background(),
		)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(communities) != 2 {
			t.Fatalf(
				"expected 2 communities, got %d",
				len(communities),
			)
		}
	})

	t.Run("returns an empty slice", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		communities, err := service.List(
			context.Background(),
		)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if communities == nil {
			t.Fatal("expected non-nil empty slice")
		}

		if len(communities) != 0 {
			t.Fatalf(
				"expected empty list, got %d entries",
				len(communities),
			)
		}
	})

	t.Run("wraps repository failure", func(t *testing.T) {
		repo := newMockCommunityRepository()
		repo.listErr = errors.New("database unavailable")

		service := NewCommunityService(repo)

		_, err := service.List(
			context.Background(),
		)

		if err == nil {
			t.Fatal("expected an error")
		}

		if !strings.Contains(err.Error(), "list communities") {
			t.Fatalf(
				"expected wrapped list error, got %v",
				err,
			)
		}
	})
}

func TestCommunityServiceUpdate(t *testing.T) {
	t.Run("updates a valid community", func(t *testing.T) {
		repo := newMockCommunityRepository()

		repo.items[1] = &model.Community{
			ID:   1,
			Name: "Old Community",
			Slug: "old-community",
		}

		service := NewCommunityService(repo)

		community := &model.Community{
			ID:             1,
			Name:           "  Updated Community  ",
			Slug:           "  UPDATED-COMMUNITY  ",
			Description:    "  Updated description.  ",
			ExternalSource: "  internal  ",
		}

		err := service.Update(
			context.Background(),
			community,
		)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if !repo.updateCalled {
			t.Fatal("expected repository Update to be called")
		}

		stored := repo.items[1]

		if stored.Name != "Updated Community" {
			t.Fatalf(
				"expected normalized name, got %q",
				stored.Name,
			)
		}

		if stored.Slug != "updated-community" {
			t.Fatalf(
				"expected normalized slug, got %q",
				stored.Slug,
			)
		}
	})

	t.Run("allows community to keep its own slug", func(t *testing.T) {
		repo := newMockCommunityRepository()

		repo.items[1] = &model.Community{
			ID:   1,
			Name: "Toronto Men",
			Slug: "toronto-men",
		}

		service := NewCommunityService(repo)

		community := &model.Community{
			ID:   1,
			Name: "Toronto Men's Community",
			Slug: "toronto-men",
		}

		err := service.Update(
			context.Background(),
			community,
		)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
	})

	t.Run("rejects duplicate slug owned by another community", func(t *testing.T) {
		repo := newMockCommunityRepository()

		repo.items[1] = &model.Community{
			ID:   1,
			Name: "Toronto Men",
			Slug: "toronto-men",
		}

		repo.items[2] = &model.Community{
			ID:   2,
			Name: "Hamilton Men",
			Slug: "hamilton-men",
		}

		service := NewCommunityService(repo)

		community := &model.Community{
			ID:   2,
			Name: "Updated Hamilton Community",
			Slug: "toronto-men",
		}

		err := service.Update(
			context.Background(),
			community,
		)

		if !errors.Is(err, ErrCommunityAlreadyExists) {
			t.Fatalf(
				"expected ErrCommunityAlreadyExists, got %v",
				err,
			)
		}

		if repo.updateCalled {
			t.Fatal("repository Update should not be called")
		}
	})

	t.Run("rejects nil community", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		err := service.Update(
			context.Background(),
			nil,
		)

		if !errors.Is(err, ErrNilCommunity) {
			t.Fatalf(
				"expected ErrNilCommunity, got %v",
				err,
			)
		}
	})

	t.Run("rejects invalid ID", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		community := validCommunity()

		err := service.Update(
			context.Background(),
			community,
		)

		if !errors.Is(err, ErrInvalidCommunityID) {
			t.Fatalf(
				"expected ErrInvalidCommunityID, got %v",
				err,
			)
		}
	})

	t.Run("preserves repository not found error", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		community := validCommunity()
		community.ID = 999

		err := service.Update(
			context.Background(),
			community,
		)

		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf(
				"expected repository.ErrNotFound, got %v",
				err,
			)
		}
	})
}

func TestCommunityServiceDelete(t *testing.T) {
	t.Run("deletes an existing community", func(t *testing.T) {
		repo := newMockCommunityRepository()

		repo.items[1] = &model.Community{
			ID:   1,
			Name: "Toronto Men",
			Slug: "toronto-men",
		}

		service := NewCommunityService(repo)

		err := service.Delete(
			context.Background(),
			1,
		)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if !repo.deleteCalled {
			t.Fatal("expected repository Delete to be called")
		}

		if _, exists := repo.items[1]; exists {
			t.Fatal("community was not deleted")
		}
	})

	t.Run("rejects invalid ID", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		err := service.Delete(
			context.Background(),
			0,
		)

		if !errors.Is(err, ErrInvalidCommunityID) {
			t.Fatalf(
				"expected ErrInvalidCommunityID, got %v",
				err,
			)
		}

		if repo.deleteCalled {
			t.Fatal("repository Delete should not be called")
		}
	})

	t.Run("preserves repository not found error", func(t *testing.T) {
		repo := newMockCommunityRepository()
		service := NewCommunityService(repo)

		err := service.Delete(
			context.Background(),
			999,
		)

		if !errors.Is(err, repository.ErrNotFound) {
			t.Fatalf(
				"expected repository.ErrNotFound, got %v",
				err,
			)
		}
	})
}
