package database

import (
	"context"
	"errors"
	"testing"
)

type mockDatabase struct {
	err error
}

func (m *mockDatabase) Health(ctx context.Context) error {
	return m.err
}

func TestHealthChecker_Name(t *testing.T) {

	checker := NewHealthChecker(
		&mockDatabase{},
	)

	if checker.Name() != "database" {
		t.Fatalf(
			"expected database got %s",
			checker.Name(),
		)
	}
}

func TestHealthChecker_Check_Success(t *testing.T) {

	checker := NewHealthChecker(
		&mockDatabase{},
	)

	err := checker.Check(context.Background())

	if err != nil {
		t.Fatalf(
			"expected nil got %v",
			err,
		)
	}
}

func TestHealthChecker_Check_Failure(t *testing.T) {

	expectedErr := errors.New("ping failed")

	checker := NewHealthChecker(
		&mockDatabase{
			err: expectedErr,
		},
	)

	err := checker.Check(context.Background())

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected %v got %v",
			expectedErr,
			err,
		)
	}
}
