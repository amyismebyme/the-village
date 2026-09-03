package external

import (
	"context"
	"errors"
	"testing"
)

func TestDeduplicateItemsKeepsFirstOccurrence(t *testing.T) {
	items := []Item{
		{
			Source:     SourceReddit,
			ExternalID: "123",
			Title:      "first",
		},
		{
			Source:     SourceReddit,
			ExternalID: "123",
			Title:      "second",
		},
		{
			Source:     SourceReddit,
			ExternalID: "456",
			Title:      "third",
		},
	}

	got, err := DeduplicateItems(
		context.Background(),
		items,
	)
	if err != nil {
		t.Fatalf(
			"deduplicate items: %v",
			err,
		)
	}

	if len(got) != 2 {
		t.Fatalf(
			"expected 2 unique items, got %d",
			len(got),
		)
	}

	if got[0].ExternalID != "123" ||
		got[0].Title != "first" {
		t.Fatalf(
			"expected first occurrence to win, got %+v",
			got[0],
		)
	}

	if got[1].ExternalID != "456" {
		t.Fatalf(
			"expected second unique identity, got %q",
			got[1].ExternalID,
		)
	}
}

func TestDeduplicateItemsTreatsDifferentSourcesAsDistinct(
	t *testing.T,
) {
	items := []Item{
		{
			Source:     SourceReddit,
			ExternalID: "123",
		},
		{
			Source:     Source("other"),
			ExternalID: "123",
		},
	}

	got, err := DeduplicateItems(
		context.Background(),
		items,
	)
	if err != nil {
		t.Fatalf(
			"deduplicate items: %v",
			err,
		)
	}

	if len(got) != 2 {
		t.Fatalf(
			"expected two distinct identities, got %d",
			len(got),
		)
	}
}

func TestDeduplicateItemsRejectsInvalidIdentity(
	t *testing.T,
) {
	items := []Item{
		{
			Source:     SourceReddit,
			ExternalID: "123",
		},
		{
			Source: SourceReddit,
		},
	}

	_, err := DeduplicateItems(
		context.Background(),
		items,
	)

	if err == nil {
		t.Fatal(
			"expected invalid identity error",
		)
	}

	if !errors.Is(
		err,
		ErrInvalidPayload,
	) {
		t.Fatalf(
			"expected ErrInvalidPayload, got %v",
			err,
		)
	}
}

func TestDeduplicateItemsHonorsCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := DeduplicateItems(
		ctx,
		[]Item{
			{
				Source:     SourceReddit,
				ExternalID: "123",
			},
		},
	)

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}
