package model

import (
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/validation"
	"strings"
	"testing"
)

func TestCommunityValidate(t *testing.T) {

	tests := []struct {
		name      string
		community Community
		wantErr   bool
	}{
		{
			name: "valid",
			community: Community{
				Name:        "Toronto Men",
				Slug:        "toronto-men",
				Description: "Helping men build meaningful friendships.",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			community: Community{
				Name: "",
				Slug: "toronto-men",
			},
			wantErr: true,
		},
		{
			name: "invalid slug",
			community: Community{
				Name: "Toronto",
				Slug: "Toronto-Men",
			},
			wantErr: true,
		},
		{
			name: "description too long",
			community: Community{
				Name:        "Toronto",
				Slug:        "toronto",
				Description: string(make([]byte, 2001)),
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {

		err := tc.community.Validate()

		if tc.wantErr && err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}

		if !tc.wantErr && err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
	}
}

func TestCommunityValidateNameBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:  "minimum valid length",
			value: "abc",
		},
		{
			name:    "below minimum",
			value:   "ab",
			wantErr: true,
		},
		{
			name:  "maximum valid length",
			value: strings.Repeat("a", 100),
		},
		{
			name:    "above maximum",
			value:   strings.Repeat("a", 101),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			community := Community{
				Name: tt.value,
				Slug: "valid-community",
			}

			err := community.Validate()

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

func TestSlugEdgeCases(t *testing.T) {
	t.Parallel()

	valid := []string{
		"a",
		"abc",
		"abc123",
		"abc-123",
		"123",
		"123-community",
	}

	for _, value := range valid {
		value := value

		t.Run("valid_"+value, func(t *testing.T) {
			t.Parallel()

			if err := validation.Slug(value); err != nil {
				t.Fatalf(
					"expected %q to be valid, got %v",
					value,
					err,
				)
			}
		})
	}

	invalid := []string{
		"A",
		"ABC",
		"abc def",
		"abc_def",
		"-abc",
		"abc-",
		"abc--def",
		"abc.def",
		"abc/def",
		"équipe",
		"你好",
	}

	for _, value := range invalid {
		value := value

		t.Run("invalid_"+value, func(t *testing.T) {
			t.Parallel()

			if err := validation.Slug(value); !errors.Is(err, validation.ErrInvalidSlug) {
				t.Fatalf(
					"expected ErrInvalidSlug for %q, got %v",
					value,
					err,
				)
			}
		})
	}
}
