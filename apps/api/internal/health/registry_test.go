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

	database := findResult(results, "database")

	if database == nil {
		t.Fatal("expected database health result")
	}

	if database.Error != "" {
		t.Fatalf(
			"expected database health check to succeed, got %q",
			database.Error,
		)
	}

	redis := findResult(results, "redis")

	if redis == nil {
		t.Fatal("expected redis health result")
	}

	if redis.Error != "" {
		t.Fatalf(
			"expected redis health check to succeed, got %q",
			redis.Error,
		)
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

	database := findResult(results, "database")

	if database == nil {
		t.Fatal("expected database health result")
	}

	if database.Error != expectedErr.Error() {
		t.Fatalf(
			"expected database error %q, got %q",
			expectedErr.Error(),
			database.Error,
		)
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

func findResult(
	results []Result,
	name string,
) *Result {
	for i := range results {
		if results[i].Name == name {
			return &results[i]
		}
	}

	return nil
}
