package external

import (
	"context"
	"testing"
)

type testProviderPayload struct {
	ID    string
	Title string
}

type testNormalizer struct{}

func (testNormalizer) Normalize(
	ctx context.Context,
	input testProviderPayload,
) (Item, error) {
	return Item{
		Source:     Source("test"),
		ExternalID: input.ID,
		Title:      input.Title,
	}, nil
}

func TestItemIdentity(t *testing.T) {
	item := Item{
		Source:     Source("reddit"),
		ExternalID: "t3_abc123",
	}

	identity := item.Identity()

	if identity.Source != item.Source {
		t.Fatalf(
			"expected source %q, got %q",
			item.Source,
			identity.Source,
		)
	}

	if identity.ExternalID != item.ExternalID {
		t.Fatalf(
			"expected external ID %q, got %q",
			item.ExternalID,
			identity.ExternalID,
		)
	}
}

func TestItemValidate(t *testing.T) {
	tests := []struct {
		name    string
		item    Item
		wantErr bool
	}{
		{
			name: "valid",
			item: Item{
				Source:     Source("reddit"),
				ExternalID: "t3_abc123",
			},
		},
		{
			name: "missing source",
			item: Item{
				ExternalID: "t3_abc123",
			},
			wantErr: true,
		},
		{
			name: "missing external id",
			item: Item{
				Source: Source("reddit"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()

			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf(
					"unexpected validation error: %v",
					err,
				)
			}
		})
	}
}

func TestNormalizerProducesProviderNeutralItem(
	t *testing.T,
) {
	normalizer := testNormalizer{}

	item, err := normalizer.Normalize(
		context.Background(),
		testProviderPayload{
			ID:    "abc123",
			Title: "Example",
		},
	)
	if err != nil {
		t.Fatalf(
			"normalize: %v",
			err,
		)
	}

	if err := item.Validate(); err != nil {
		t.Fatalf(
			"validate normalized item: %v",
			err,
		)
	}

	if item.Source != Source("test") {
		t.Fatalf(
			"unexpected source %q",
			item.Source,
		)
	}

	if item.ExternalID != "abc123" {
		t.Fatalf(
			"unexpected external ID %q",
			item.ExternalID,
		)

	}
}
