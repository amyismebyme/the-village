package testutil

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScriptedHTTPClientConsumesOutcomesInOrder(
	t *testing.T,
) {
	base := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(http.StatusOK)

				_, _ = io.WriteString(
					w,
					"base",
				)
			},
		),
	)

	defer base.Close()

	client := NewScriptedHTTPClient(
		base.Client(),
	)

	client.SetScript(
		"/dependency",
		HTTPOutcome{
			StatusCode: http.StatusServiceUnavailable,
		},
		HTTPOutcome{
			StatusCode: http.StatusOK,
			Body:       "recovered",
		},
	)

	for i, wantStatus := range []int{
		http.StatusServiceUnavailable,
		http.StatusOK,
	} {
		req, err := http.NewRequest(
			http.MethodGet,
			base.URL+"/dependency",
			nil,
		)
		if err != nil {
			t.Fatalf(
				"create request %d: %v",
				i+1,
				err,
			)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf(
				"request %d: %v",
				i+1,
				err,
			)
		}

		_ = resp.Body.Close()

		if resp.StatusCode != wantStatus {
			t.Fatalf(
				"request %d: expected status %d, got %d",
				i+1,
				wantStatus,
				resp.StatusCode,
			)
		}
	}

	if got := client.Calls(
		"/dependency",
	); got != 2 {
		t.Fatalf(
			"expected two scripted calls, got %d",
			got,
		)
	}
}

func TestScriptedHTTPClientFallsBackToBase(
	t *testing.T,
) {
	base := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(http.StatusOK)

				_, _ = io.WriteString(
					w,
					"base",
				)
			},
		),
	)

	defer base.Close()

	client := NewScriptedHTTPClient(
		base.Client(),
	)

	client.SetScript(
		"/dependency",
		HTTPOutcome{
			StatusCode: http.StatusBadGateway,
		},
	)

	req, err := http.NewRequest(
		http.MethodGet,
		base.URL+"/dependency",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"create request: %v",
			err,
		)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf(
			"first request: %v",
			err,
		)
	}

	_ = resp.Body.Close()

	req, err = http.NewRequest(
		http.MethodGet,
		base.URL+"/dependency",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"create second request: %v",
			err,
		)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf(
			"fallback request: %v",
			err,
		)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(
			"read fallback body: %v",
			err,
		)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected fallback status 200, got %d",
			resp.StatusCode,
		)
	}

	if string(body) != "base" {
		t.Fatalf(
			"expected fallback body %q, got %q",
			"base",
			string(body),
		)
	}
}

func TestScriptedHTTPClientCanInjectTransportError(
	t *testing.T,
) {
	base := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	defer base.Close()

	expected := errors.New(
		"synthetic transport failure",
	)

	client := NewScriptedHTTPClient(
		base.Client(),
	)

	client.SetScript(
		"/dependency",
		HTTPOutcome{
			Err: expected,
		},
	)

	req, err := http.NewRequest(
		http.MethodGet,
		base.URL+"/dependency",
		nil,
	)
	if err != nil {
		t.Fatalf(
			"create request: %v",
			err,
		)
	}

	_, err = client.Do(req)

	if !errors.Is(err, expected) {
		t.Fatalf(
			"expected injected error, got %v",
			err,
		)
	}
}
