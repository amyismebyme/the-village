package service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"testing"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type mockCommunityRepository struct {
	items map[int64]*model.Community
}

func newMockRepository() *mockCommunityRepository {

	return &mockCommunityRepository{
		items: make(map[int64]*model.Community),
	}
}

func (m *mockCommunityRepository) Create(
	ctx context.Context,
	community *model.Community,
) error {

	m.items[community.ID] = community
	return nil
}

func (m *mockCommunityRepository) FindByID(
	ctx context.Context,
	id int64,
) (*model.Community, error) {

	c, ok := m.items[id]

	if !ok {
		return nil, errors.New("not found")
	}

	return c, nil
}

func (m *mockCommunityRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (*model.Community, error) {

	for _, c := range m.items {

		if c.Slug == slug {
			return c, nil
		}
	}

	return nil, pgx.ErrNoRows
}

func (m *mockCommunityRepository) List(
	ctx context.Context,
) ([]*model.Community, error) {

	var communities []*model.Community

	for _, c := range m.items {
		communities = append(communities, c)
	}

	return communities, nil
}

func (m *mockCommunityRepository) Update(
	ctx context.Context,
	community *model.Community,
) error {

	m.items[community.ID] = community
	return nil
}

func (m *mockCommunityRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	delete(m.items, id)
	return nil
}

func TestCreateCommunity(t *testing.T) {

	repo := newMockRepository()

	service := NewCommunityService(repo)

	community := &model.Community{
		ID:   1,
		Name: "Toronto Men",
		Slug: "toronto-men",
	}

	err := service.Create(context.Background(), community)

	if err != nil {
		t.Fatal(err)
	}
}

func TestDuplicateSlug(t *testing.T) {

	repo := newMockRepository()

	repo.items[1] = &model.Community{
		ID:   1,
		Name: "Existing",
		Slug: "toronto-men",
	}

	service := NewCommunityService(repo)

	err := service.Create(context.Background(), &model.Community{
		ID:   2,
		Name: "New",
		Slug: "toronto-men",
	})

	if err == nil {
		t.Fatal("expected duplicate slug error")
	}
}

func TestGetCommunity(t *testing.T) {

	repo := newMockRepository()

	repo.items[10] = &model.Community{
		ID:   10,
		Name: "Toronto",
		Slug: "toronto",
	}

	service := NewCommunityService(repo)

	c, err := service.Get(context.Background(), 10)

	if err != nil {
		t.Fatal(err)
	}

	if c.Name != "Toronto" {
		t.Fatal("unexpected name")
	}
}

func TestUpdateCommunity(t *testing.T) {

	repo := newMockRepository()

	repo.items[1] = &model.Community{
		ID:   1,
		Name: "Old",
		Slug: "old",
	}

	service := NewCommunityService(repo)

	updated := &model.Community{
		ID:   1,
		Name: "New",
		Slug: "new",
	}

	if err := service.Update(context.Background(), updated); err != nil {
		t.Fatal(err)
	}

	if repo.items[1].Name != "New" {
		t.Fatal("update failed")
	}
}

func TestDeleteCommunity(t *testing.T) {

	repo := newMockRepository()

	repo.items[1] = &model.Community{
		ID: 1,
	}

	service := NewCommunityService(repo)

	if err := service.Delete(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	if len(repo.items) != 0 {
		t.Fatal("delete failed")
	}
}

func (m *mockCommunityRepository) DeleteAll(ctx context.Context) error {
	for k := range m.items {
		delete(m.items, k)
	}
	return nil
}
