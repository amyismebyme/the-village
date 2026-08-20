package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{
			name:    "empty",
			value:   "",
			wantErr: ErrRequired,
		},
		{
			name:    "spaces",
			value:   "   ",
			wantErr: ErrRequired,
		},
		{
			name:    "tabs",
			value:   "\t\t",
			wantErr: ErrRequired,
		},
		{
			name:  "valid",
			value: "hello",
		},
		{
			name:  "unicode",
			value: "你好",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Required(tt.value)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf(
						"expected valid value, got %v",
						err,
					)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}

func TestLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		min     int
		max     int
		wantErr error
	}{
		{
			name:    "below minimum",
			value:   "ab",
			min:     3,
			max:     10,
			wantErr: ErrTooShort,
		},
		{
			name:  "exact minimum",
			value: "abc",
			min:   3,
			max:   10,
		},
		{
			name:  "middle",
			value: "abcdef",
			min:   3,
			max:   10,
		},
		{
			name:  "exact maximum",
			value: "abcdefghij",
			min:   3,
			max:   10,
		},
		{
			name:    "above maximum",
			value:   "abcdefghijk",
			min:     3,
			max:     10,
			wantErr: ErrTooLong,
		},
		{
			name:    "invalid minimum",
			value:   "abc",
			min:     -1,
			max:     10,
			wantErr: ErrInvalidLength,
		},
		{
			name:    "minimum greater than maximum",
			value:   "abc",
			min:     10,
			max:     3,
			wantErr: ErrInvalidLength,
		},
		{
			name:  "unicode uses rune length",
			value: "你好",
			min:   2,
			max:   2,
		},
		{
			name:  "surrounding whitespace is ignored",
			value: "  hello  ",
			min:   5,
			max:   5,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Length(
				tt.value,
				tt.min,
				tt.max,
			)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf(
						"expected valid value, got %v",
						err,
					)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		max     int
		wantErr error
	}{
		{
			name:  "below maximum",
			value: "abc",
			max:   5,
		},
		{
			name:  "exact maximum",
			value: "abcde",
			max:   5,
		},
		{
			name:    "above maximum",
			value:   "abcdef",
			max:     5,
			wantErr: ErrTooLong,
		},
		{
			name:    "invalid maximum",
			value:   "abc",
			max:     -1,
			wantErr: ErrInvalidLength,
		},
		{
			name:  "unicode uses rune length",
			value: "你好",
			max:   2,
		},
		{
			name:  "surrounding whitespace is ignored",
			value: "  hello  ",
			max:   5,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := MaxLength(
				tt.value,
				tt.max,
			)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf(
						"expected valid value, got %v",
						err,
					)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}

func TestSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "simple",
			value: "vancouver",
			valid: true,
		},
		{
			name:  "hyphenated",
			value: "toronto-men",
			valid: true,
		},
		{
			name:  "numbers",
			value: "community1",
			valid: true,
		},
		{
			name:  "mixed numbers",
			value: "men-123",
			valid: true,
		},
		{
			name:  "empty",
			value: "",
		},
		{
			name:  "uppercase",
			value: "Toronto",
		},
		{
			name:  "uppercase mixed",
			value: "Toronto-Men",
		},
		{
			name:  "underscore",
			value: "hello_world",
		},
		{
			name:  "space",
			value: "hello world",
		},
		{
			name:  "leading hyphen",
			value: "-hello",
		},
		{
			name:  "trailing hyphen",
			value: "hello-",
		},
		{
			name:  "double hyphen",
			value: "hello--world",
		},
		{
			name:  "special characters",
			value: "hello@world",
		},
		{
			name:  "unicode",
			value: "toronto-men-你好",
		},
		{
			name:  "whitespace",
			value: " hello ",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Slug(tt.value)

			if tt.valid {
				if err != nil {
					t.Fatalf(
						"expected valid slug, got %v",
						err,
					)
				}

				return
			}

			if err == nil {
				t.Fatalf(
					"expected %q to be invalid",
					tt.value,
				)
			}

			if !errors.Is(err, ErrInvalidSlug) {
				t.Fatalf(
					"expected ErrInvalidSlug, got %v",
					err,
				)
			}
		})
	}
}

func TestLengthUsesRuneCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		min     int
		max     int
		wantErr error
	}{
		{
			name:  "two chinese characters",
			value: "你好",
			min:   2,
			max:   2,
		},
		{
			name:    "two chinese characters below minimum",
			value:   "你好",
			min:     3,
			max:     5,
			wantErr: ErrTooShort,
		},
		{
			name:    "two chinese characters above maximum",
			value:   "你好",
			min:     1,
			max:     1,
			wantErr: ErrTooLong,
		},
		{
			name:  "accented characters",
			value: "éàö",
			min:   3,
			max:   3,
		},
		{
			name:  "emoji",
			value: "😀😀",
			min:   2,
			max:   2,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Length(
				tt.value,
				tt.min,
				tt.max,
			)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf(
						"expected valid value, got %v",
						err,
					)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}

func TestMaxLengthBoundary(t *testing.T) {
	t.Parallel()

	exact := strings.Repeat("a", 100)
	above := strings.Repeat("a", 101)

	if err := MaxLength(exact, 100); err != nil {
		t.Fatalf(
			"expected exact max length to be valid, got %v",
			err,
		)
	}

	if err := MaxLength(above, 100); !errors.Is(err, ErrTooLong) {
		t.Fatalf(
			"expected ErrTooLong, got %v",
			err,
		)
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

			if err := Slug(value); err != nil {
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

			if err := Slug(value); !errors.Is(err, ErrInvalidSlug) {
				t.Fatalf(
					"expected ErrInvalidSlug for %q, got %v",
					value,
					err,
				)
			}
		})
	}
}
