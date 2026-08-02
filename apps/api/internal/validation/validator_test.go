package validation

import "testing"

func TestRequired(t *testing.T) {

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{"empty", "", false},
		{"spaces", "   ", false},
		{"valid", "hello", true},
	}

	for _, tt := range tests {

		err := Required(tt.value)

		if tt.valid && err != nil {
			t.Fatalf("%s should be valid", tt.name)
		}

		if !tt.valid && err == nil {
			t.Fatalf("%s should fail", tt.name)
		}
	}
}

func TestLength(t *testing.T) {

	if Length("ab", 3, 10) == nil {
		t.Fatal("expected short string to fail")
	}

	if Length("abcdefghijk", 3, 10) == nil {
		t.Fatal("expected long string to fail")
	}

	if err := Length("abcdef", 3, 10); err != nil {
		t.Fatal(err)
	}
}

func TestMaxLength(t *testing.T) {

	if MaxLength("abcdef", 5) == nil {
		t.Fatal("expected failure")
	}

	if err := MaxLength("abc", 5); err != nil {
		t.Fatal(err)
	}
}

func TestSlug(t *testing.T) {

	valid := []string{
		"toronto-men",
		"vancouver",
		"village-community",
		"community1",
	}

	for _, s := range valid {

		if err := Slug(s); err != nil {
			t.Fatalf("%s should be valid", s)
		}
	}

	invalid := []string{
		"",
		"Toronto",
		"Toronto-Men",
		"hello_world",
		"hello world",
		"-hello",
		"hello-",
	}

	for _, s := range invalid {

		if Slug(s) == nil {
			t.Fatalf("%s should be invalid", s)
		}
	}
}
