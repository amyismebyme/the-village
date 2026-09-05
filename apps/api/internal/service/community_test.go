package service

import (
	"context"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/repository"
	"github.com/amyismebyme/the-village/apps/api/internal/validation"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"testing"
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
	limit int,
	offset int,
) ([]*model.Community, int64, error) {

	if m.listErr != nil {
		return nil, 0, m.listErr
	}

	list := make([]*model.Community, 0, len(m.communities))

	for _, c := range m.communities {

		copy := *c
		list = append(list, &copy)
	}

	total := int64(len(list))

	if offset >= len(list) {
		return []*model.Community{}, total, nil
	}

	end := offset + limit
	if end > len(list) {
		end = len(list)
	}

	return list[offset:end], total, nil
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

	result, err := svc.List(
		context.Background(),
		20,
		0,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(result.Communities) != 2 {
		t.Fatalf("expected 2 communities, got %d", len(result.Communities))
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

func TestCommunityServiceCreateInvalidCommunityDoesNotReachRepository(
	t *testing.T,
) {
	t.Parallel()

	repo := newMockCommunityRepository()

	community := &model.Community{
		Name: "",
		Slug: "",
	}

	err := NewCommunityService(repo).Create(
		context.Background(),
		community,
	)

	if !errors.Is(
		err,
		ErrInvalidCommunity,
	) {
		t.Fatalf(
			"expected ErrInvalidCommunity, got %v",
			err,
		)
	}

	if len(repo.communities) != 0 {
		t.Fatalf(
			"expected repository to remain empty, got %d records",
			len(repo.communities),
		)
	}
}

func TestCommunityServiceUpdateInvalidCommunityDoesNotReachRepository(
	t *testing.T,
) {
	t.Parallel()

	repo := newMockCommunityRepository()

	community := &model.Community{
		ID:   1,
		Name: "",
		Slug: "",
	}

	err := NewCommunityService(repo).Update(
		context.Background(),
		community,
	)

	if !errors.Is(
		err,
		ErrInvalidCommunity,
	) {
		t.Fatalf(
			"expected ErrInvalidCommunity, got %v",
			err,
		)
	}
}

func TestCommunityServiceCreatePreservesSlugValidationCause(
	t *testing.T,
) {
	t.Parallel()

	repo := newMockCommunityRepository()

	service := NewCommunityService(repo)

	err := service.Create(
		context.Background(),
		&model.Community{
			Name: "Valid Community",
			Slug: "Invalid Slug",
		},
	)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if !errors.Is(
		err,
		ErrInvalidCommunity,
	) {
		t.Fatalf(
			"expected ErrInvalidCommunity, got %v",
			err,
		)
	}

	if !errors.Is(
		err,
		validation.ErrInvalidSlug,
	) {
		t.Fatalf(
			"expected validation.ErrInvalidSlug to be preserved, got %v",
			err,
		)
	}
}

func TestCreateCommunityMetricSuccess(t *testing.T) {
	repo := newMockCommunityRepository()
	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	community := &model.Community{
		Name: "Toronto Men",
		Slug: "toronto-men",
	}

	if err := svc.Create(
		context.Background(),
		community,
	); err != nil {
		t.Fatalf(
			"Create failed: %v",
			err,
		)
	}

	after := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected create metric +1, got %v",
			got,
		)
	}
}

func TestCreateCommunityMetricFailureIncrementsFailure(t *testing.T) {
	repo := newMockCommunityRepository()
	repo.createErr = errors.New("database unavailable")

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityCreateTotal.WithLabelValues("failure"),
	)

	err := svc.Create(
		context.Background(),
		&model.Community{
			Name: "Toronto Men",
			Slug: "toronto-men",
		},
	)
	if err == nil {
		t.Fatal("expected Create error")
	}

	after := testutil.ToFloat64(
		metrics.CommunityCreateTotal.WithLabelValues("failure"),
	)

	if got := after - before; got != 1 {
		t.Fatalf("expected create failure metric +1, got %v", got)
	}
}

func TestCreateCommunityMetricFailureDoesNotIncrementSuccess(
	t *testing.T,
) {
	repo := newMockCommunityRepository()

	repo.createErr = errors.New(
		"database unavailable",
	)

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	err := svc.Create(
		context.Background(),
		&model.Community{
			Name: "Toronto Men",
			Slug: "toronto-men",
		},
	)

	if err == nil {
		t.Fatal("expected Create error")
	}

	after := testutil.ToFloat64(
		metrics.CommunityCreateTotal.
			WithLabelValues("success"),
	)

	if got := after - before; got != 0 {
		t.Fatalf(
			"expected create success metric unchanged, got %v",
			got,
		)
	}
}

func TestUpdateCommunityMetricSuccess(t *testing.T) {
	repo := newMockCommunityRepository()

	repo.communities[1] = &model.Community{
		ID:   1,
		Name: "Toronto Men",
		Slug: "toronto-men",
	}

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.
			WithLabelValues("success"),
	)

	community := &model.Community{
		ID:   1,
		Name: "Toronto Men Updated",
		Slug: "toronto-men-updated",
	}

	if err := svc.Update(
		context.Background(),
		community,
	); err != nil {
		t.Fatalf(
			"Update failed: %v",
			err,
		)
	}

	after := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.
			WithLabelValues("success"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected update metric +1, got %v",
			got,
		)
	}
}

func TestUpdateCommunityMetricFailureIncrementsFailure(t *testing.T) {
	repo := newMockCommunityRepository()
	repo.communities[1] = &model.Community{
		ID:   1,
		Name: "Toronto Men",
		Slug: "toronto-men",
	}
	repo.updateErr = errors.New("database unavailable")

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.WithLabelValues("failure"),
	)

	err := svc.Update(
		context.Background(),
		&model.Community{
			ID:   1,
			Name: "Toronto Men Updated",
			Slug: "toronto-men-updated",
		},
	)
	if err == nil {
		t.Fatal("expected Update error")
	}

	after := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.WithLabelValues("failure"),
	)

	if got := after - before; got != 1 {
		t.Fatalf("expected update failure metric +1, got %v", got)
	}
}

func TestUpdateCommunityMetricFailureDoesNotIncrementSuccess(
	t *testing.T,
) {
	repo := newMockCommunityRepository()

	repo.communities[1] = &model.Community{
		ID:   1,
		Name: "Toronto Men",
		Slug: "toronto-men",
	}

	repo.updateErr = errors.New(
		"database unavailable",
	)

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.
			WithLabelValues("success"),
	)

	err := svc.Update(
		context.Background(),
		&model.Community{
			ID:   1,
			Name: "Toronto Men Updated",
			Slug: "toronto-men-updated",
		},
	)

	if err == nil {
		t.Fatal("expected Update error")
	}

	after := testutil.ToFloat64(
		metrics.CommunityUpdateTotal.
			WithLabelValues("success"),
	)

	if got := after - before; got != 0 {
		t.Fatalf(
			"expected update success metric unchanged, got %v",
			got,
		)
	}
}

func TestDeleteCommunityMetricSuccess(t *testing.T) {
	repo := newMockCommunityRepository()

	repo.communities[1] = &model.Community{
		ID:   1,
		Name: "Toronto Men",
		Slug: "toronto-men",
	}

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	if err := svc.Delete(
		context.Background(),
		1,
	); err != nil {
		t.Fatalf(
			"Delete failed: %v",
			err,
		)
	}

	after := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected delete metric +1, got %v",
			got,
		)
	}
}

func TestDeleteCommunityMetricNotFoundDoesNotIncrementSuccess(
	t *testing.T,
) {
	repo := newMockCommunityRepository()
	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	err := svc.Delete(
		context.Background(),
		999,
	)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf(
			"expected repository.ErrNotFound, got %v",
			err,
		)
	}

	after := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	if got := after - before; got != 0 {
		t.Fatalf(
			"expected delete success metric unchanged, got %v",
			got,
		)
	}
}

func TestDeleteCommunityMetricFailureIncrementsFailure(t *testing.T) {
	repo := newMockCommunityRepository()
	repo.communities[1] = &model.Community{
		ID:   1,
		Name: "Toronto Men",
		Slug: "toronto-men",
	}
	repo.deleteErr = errors.New("database unavailable")

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.WithLabelValues("failure"),
	)

	err := svc.Delete(
		context.Background(),
		1,
	)
	if err == nil {
		t.Fatal("expected Delete error")
	}

	after := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.WithLabelValues("failure"),
	)

	if got := after - before; got != 1 {
		t.Fatalf("expected delete failure metric +1, got %v", got)
	}
}

func TestDeleteCommunityMetricFailureDoesNotIncrementSuccess(
	t *testing.T,
) {
	repo := newMockCommunityRepository()

	repo.communities[1] = &model.Community{
		ID:   1,
		Name: "Toronto Men",
		Slug: "toronto-men",
	}

	repo.deleteErr = errors.New(
		"database unavailable",
	)

	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	err := svc.Delete(
		context.Background(),
		1,
	)

	if err == nil {
		t.Fatal("expected Delete error")
	}

	after := testutil.ToFloat64(
		metrics.CommunityDeleteTotal.
			WithLabelValues("success"),
	)

	if got := after - before; got != 0 {
		t.Fatalf(
			"expected delete success metric unchanged, got %v",
			got,
		)
	}
}

func TestCommunityValidationFailureMetricName(t *testing.T) {
	repo := newMockCommunityRepository()
	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("name"),
	)

	err := svc.Create(
		context.Background(),
		&model.Community{
			Name: "",
			Slug: "valid-community",
		},
	)

	if !errors.Is(err, ErrInvalidCommunity) {
		t.Fatalf(
			"expected ErrInvalidCommunity, got %v",
			err,
		)
	}

	after := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("name"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected name validation metric +1, got %v",
			got,
		)
	}
}

func TestCommunityValidationFailureMetricSlug(t *testing.T) {
	repo := newMockCommunityRepository()
	svc := NewCommunityService(repo)

	before := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("slug"),
	)

	err := svc.Create(
		context.Background(),
		&model.Community{
			Name: "Toronto Men",
			Slug: "Invalid Slug",
		},
	)

	if !errors.Is(err, ErrInvalidCommunity) {
		t.Fatalf(
			"expected ErrInvalidCommunity, got %v",
			err,
		)
	}

	after := testutil.ToFloat64(
		metrics.CommunityValidationFailuresTotal.
			WithLabelValues("slug"),
	)

	if got := after - before; got != 1 {
		t.Fatalf(
			"expected slug validation metric +1, got %v",
			got,
		)
	}
}

func TestCommunityValidationMetricDoesNotExposeValidationMessage(
	t *testing.T,
) {
	registry := prometheus.NewRegistry()

	collector := metrics.CommunityValidationFailuresTotal

	if err := registry.Register(collector); err != nil {
		t.Fatalf(
			"register validation metric: %v",
			err,
		)
	}

	collector.
		WithLabelValues("name").
		Inc()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather validation metric: %v",
			err,
		)
	}

	for _, family := range families {
		if family.GetName() !=
			"village_community_validation_failures_total" {
			continue
		}

		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() != "field" {
					t.Fatalf(
						"unexpected validation metric label %q",
						label.GetName(),
					)
				}

				if label.GetValue() == "name: value is too short" {
					t.Fatalf(
						"raw validation message leaked into metric label",
					)
				}
			}
		}
	}
}

func TestFieldErrorPreservesCause(t *testing.T) {
	err := validation.NewFieldError(
		"name",
		validation.ErrTooShort,
	)

	if !errors.Is(
		err,
		validation.ErrTooShort,
	) {
		t.Fatal(
			"expected validation cause to be preserved",
		)
	}
}

func TestFieldErrorExposesField(t *testing.T) {
	err := validation.NewFieldError(
		"slug",
		validation.ErrInvalidSlug,
	)

	var fieldErr validation.FieldError

	if !errors.As(err, &fieldErr) {
		t.Fatal(
			"expected FieldError",
		)
	}

	if fieldErr.Field != "slug" {
		t.Fatalf(
			"expected field slug, got %q",
			fieldErr.Field,
		)
	}
}
