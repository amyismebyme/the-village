package health

import (
	"context"
	"errors"
	"testing"
)

type mockChecker struct {
	name string
	err  error
}

func (m mockChecker) Name() string {
	return m.name
}

func (m mockChecker) Check(ctx context.Context) error {
	return m.err
}

func TestRegistry_Register(t *testing.T) {

	registry := NewRegistry()

	registry.Register(mockChecker{
		name: "database",
	})

	if len(registry.checkers) != 1 {
		t.Fatalf(
			"expected 1 checker, got %d",
			len(registry.checkers),
		)
	}
}

func TestRegistry_Check_AllHealthy(t *testing.T) {

	registry := NewRegistry()

	registry.Register(mockChecker{
		name: "database",
	})

	registry.Register(mockChecker{
		name: "redis",
	})

	results := registry.Check(context.Background())

	if len(results) != 2 {
		t.Fatalf(
			"expected 2 results, got %d",
			len(results),
		)
	}

	if results["database"] != nil {
		t.Fatal("expected database health check to succeed")
	}

	if results["redis"] != nil {
		t.Fatal("expected redis health check to succeed")
	}
}

func TestRegistry_Check_WithFailure(t *testing.T) {

	expectedErr := errors.New("database unavailable")

	registry := NewRegistry()

	registry.Register(mockChecker{
		name: "database",
		err:  expectedErr,
	})

	results := registry.Check(context.Background())

	if !errors.Is(results["database"], expectedErr) {
		t.Fatal("expected database error")
	}
}

func TestRegistry_Check_EmptyRegistry(t *testing.T) {

	registry := NewRegistry()

	results := registry.Check(context.Background())

	if len(results) != 0 {
		t.Fatalf(
			"expected empty results, got %d",
			len(results),
		)
	}
}
