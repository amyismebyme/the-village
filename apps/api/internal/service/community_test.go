package service

import (
	"context"
	"errors"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
)

type mockCommunityRepository struct {
	communities map[int64]*model.Community
	nextID      int64

	findBySlugErr error
	createErr     error
	updateErr     error
	deleteErr     error
	findErr       error
	listErr       error
}

func newMockCommunityRepository() *mockCommunityRepository {
	return &mockCommunityRepository{
		communities: make(map[int64]*model.Community),
		nextID:      1,
	}
}

func (m *mockCommunityRepository) Create(
	ctx context.Context,
	community *model.Community,
) error {

	if m.createErr != nil {
		return m.createErr
	}

	community.ID = m.nextID
	m.nextID++

	copy := *community
	m.communities[community.ID] = &copy

	return nil
}

func (m *mockCommunityRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.Community, error) {

	if m.findErr != nil {
		return nil, m.findErr
	}

	c, ok := m.communities[id]

	if !ok {
		return nil, repository.ErrNotFound
	}

	copy := *c
	return &copy, nil
}

func (m *mockCommunityRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (*model.Community, error) {

	if m.findBySlugErr != nil {
		return nil, m.findBySlugErr
	}

	for _, c := range m.communities {

		if c.Slug == slug {

			copy := *c
			return &copy, nil
		}
	}

	return nil, repository.ErrNotFound
}

func (m *mockCommunityRepository) List(
	ctx context.Context,
) ([]*model.Community, error) {

	if m.listErr != nil {
		return nil, m.listErr
	}

	list := make([]*model.Community, 0, len(m.communities))

	for _, c := range m.communities {

		copy := *c
		list = append(list, &copy)
	}

	return list, nil
}

func (m *mockCommunityRepository) Update(
	ctx context.Context,
	community *model.Community,
) error {

	if m.updateErr != nil {
		return m.updateErr
	}

	if _, ok := m.communities[community.ID]; !ok {
		return repository.ErrNotFound
	}

	copy := *community
	m.communities[community.ID] = &copy

	return nil
}

func (m *mockCommunityRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	if m.deleteErr != nil {
		return m.deleteErr
	}

	if _, ok := m.communities[id]; !ok {
		return repository.ErrNotFound
	}

	delete(m.communities, id)

	return nil
}

func TestCreateCommunity(t *testing.T) {

	repo := newMockCommunityRepository()

	svc := NewCommunityService(repo)

	community := &model.Community{
		Name:        "Toronto Men's Group",
		Slug:        "Toronto-Men",
		Description: "Helping men connect",
	}

	if err := svc.Create(context.Background(), community); err != nil {
		t.Fatal(err)
	}

	if community.ID == 0 {
		t.Fatal("expected repository to assign an ID")
	}

	if community.Slug != "toronto-men" {
		t.Fatalf("expected normalized slug, got %q", community.Slug)
	}
}

func TestDuplicateSlug(t *testing.T) {

	repo := newMockCommunityRepository()

	existing := &model.Community{
		Name: "Existing",
		Slug: "toronto",
	}

	if err := repo.Create(context.Background(), existing); err != nil {
		t.Fatal(err)
	}

	svc := NewCommunityService(repo)

	err := svc.Create(
		context.Background(),
		&model.Community{
			Name: "Another",
			Slug: "toronto",
		},
	)

	if !errors.Is(err, ErrCommunityAlreadyExists) {
		t.Fatalf("expected duplicate slug error, got %v", err)
	}
}

func TestGetCommunity(t *testing.T) {

	repo := newMockCommunityRepository()

	community := &model.Community{
		Name: "Toronto",
		Slug: "toronto",
	}

	_ = repo.Create(context.Background(), community)

	svc := NewCommunityService(repo)

	got, err := svc.Get(context.Background(), community.ID)

	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "Toronto" {
		t.Fatal("unexpected community returned")
	}
}

func TestListCommunities(t *testing.T) {

	repo := newMockCommunityRepository()

	_ = repo.Create(context.Background(), &model.Community{
		Name: "One",
		Slug: "one",
	})

	_ = repo.Create(context.Background(), &model.Community{
		Name: "Two",
		Slug: "two",
	})

	svc := NewCommunityService(repo)

	list, err := svc.List(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 communities, got %d", len(list))
	}
}

func TestUpdateCommunity(t *testing.T) {

	repo := newMockCommunityRepository()

	community := &model.Community{
		Name: "Old",
		Slug: "old",
	}

	_ = repo.Create(context.Background(), community)

	svc := NewCommunityService(repo)

	community.Name = "New"

	if err := svc.Update(context.Background(), community); err != nil {
		t.Fatal(err)
	}

	got, _ := repo.FindByID(context.Background(), community.ID)

	if got.Name != "New" {
		t.Fatal("update failed")
	}
}

func TestDeleteCommunity(t *testing.T) {

	repo := newMockCommunityRepository()

	community := &model.Community{
		Name: "Delete Me",
		Slug: "delete-me",
	}

	_ = repo.Create(context.Background(), community)

	svc := NewCommunityService(repo)

	if err := svc.Delete(context.Background(), community.ID); err != nil {
		t.Fatal(err)
	}

	_, err := repo.FindByID(context.Background(), community.ID)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatal("community should have been deleted")
	}
}
