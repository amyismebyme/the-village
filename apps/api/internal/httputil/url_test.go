package httputil

import "testing"

func TestParseBaseURLAllowsHTTPWhenRequested(
	t *testing.T,
) {
	got, err := ParseBaseURL(
		"http://localhost:8080",
		true,
	)
	if err != nil {
		t.Fatalf(
			"expected valid URL, got %v",
			err,
		)
	}

	if got.Host != "localhost:8080" {
		t.Fatalf(
			"unexpected host %q",
			got.Host,
		)
	}
}

func TestParseBaseURLRejectsHTTPWhenHTTPSRequired(
	t *testing.T,
) {
	_, err := ParseBaseURL(
		"http://example.com",
		false,
	)

	if err == nil {
		t.Fatal(
			"expected HTTP URL to be rejected",
		)
	}
}

func TestParseBaseURLRejectsMissingHost(
	t *testing.T,
) {
	_, err := ParseBaseURL(
		"https:///missing-host",
		false,
	)

	if err == nil {
		t.Fatal(
			"expected missing host error",
		)
	}
}

func TestParseBaseURLRejectsEmptyValue(
	t *testing.T,
) {
	_, err := ParseBaseURL(
		"",
		true,
	)

	if err == nil {
		t.Fatal(
			"expected empty URL error",
		)
	}
}
