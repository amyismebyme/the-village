//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
)

func TestRepositoryContextCancellation(t *testing.T) {
	app := newCommunityRepositoryTestApp(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := app.repo.Create(
		ctx,
		newTestCommunity(),
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}
