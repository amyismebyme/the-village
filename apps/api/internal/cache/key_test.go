package cache

import "testing"

func TestCommunityKey(t *testing.T) {
	got, err := CommunityKey(
		"api",
		"17",
	)
	if err != nil {
		t.Fatalf(
			"create key: %v",
			err,
		)
	}

	want := "community:api:lookup:17"

	if got != want {
		t.Fatalf(
			"expected %q, got %q",
			want,
			got,
		)
	}
}

func TestRedditListingKey(t *testing.T) {
	got, err := RedditListingKey(
		"toronto",
		"abc123",
		25,
	)
	if err != nil {
		t.Fatalf(
			"create key: %v",
			err,
		)
	}

	want :=
		"reddit:reddit:listing:subreddit=toronto,after=abc123,limit=25"

	if got != want {
		t.Fatalf(
			"expected %q, got %q",
			want,
			got,
		)
	}
}

func TestRedditListingKeyChangesWhenPagingChanges(
	t *testing.T,
) {
	first, err := RedditListingKey(
		"toronto",
		"",
		25,
	)
	if err != nil {
		t.Fatalf(
			"first key: %v",
			err,
		)
	}

	second, err := RedditListingKey(
		"toronto",
		"abc123",
		25,
	)
	if err != nil {
		t.Fatalf(
			"second key: %v",
			err,
		)
	}

	if first == second {
		t.Fatal(
			"expected different cache keys for different paging state",
		)
	}
}

func TestRedditListingKeyRejectsInvalidInput(
	t *testing.T,
) {
	tests := []struct {
		name      string
		subreddit string
		limit     int
	}{
		{
			name:      "missing subreddit",
			subreddit: "",
			limit:     25,
		},
		{
			name:      "zero limit",
			subreddit: "toronto",
			limit:     0,
		},
		{
			name:      "negative limit",
			subreddit: "toronto",
			limit:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RedditListingKey(
				tt.subreddit,
				"",
				tt.limit,
			)

			if err == nil {
				t.Fatal(
					"expected error",
				)
			}
		})
	}
}
