package reddit

import (
	"bytes"
	"errors"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestRedditObservabilityRecordsRequest(
	t *testing.T,
) {
	metrics.ExternalRequestsTotal.Reset()
	metrics.ExternalRequestDuration.Reset()
	metrics.ExternalErrorsTotal.Reset()

	registry := prometheus.NewRegistry()

	for _, collector := range []prometheus.Collector{
		metrics.ExternalRequestsTotal,
		metrics.ExternalRequestDuration,
		metrics.ExternalErrorsTotal,
	} {
		if err := registry.Register(collector); err != nil {
			t.Fatalf(
				"register collector: %v",
				err,
			)
		}
	}

	observeOperation(
		nil,
		"fetch",
		"",
		"200",
		true,
		time.Now(),
		nil,
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	assertMetricFamilyPresent(
		t,
		families,
		"village_external_requests_total",
	)

	assertMetricFamilyPresent(
		t,
		families,
		"village_external_request_duration_seconds",
	)
}

func TestRedditObservabilityRecordsRateLimitError(
	t *testing.T,
) {
	metrics.ExternalRequestsTotal.Reset()
	metrics.ExternalRequestDuration.Reset()
	metrics.ExternalErrorsTotal.Reset()

	registry := prometheus.NewRegistry()

	registry.MustRegister(
		metrics.ExternalRequestsTotal,
		metrics.ExternalRequestDuration,
		metrics.ExternalErrorsTotal,
	)

	observeOperation(
		nil,
		"fetch",
		"",
		"200",
		true,
		time.Now(),
		external.ErrRateLimited,
	)

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf(
			"gather metrics: %v",
			err,
		)
	}

	assertMetricFamilyPresent(
		t,
		families,
		"village_external_errors_total",
	)
}

func TestRedditErrorStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "unauthorized",
			err:  external.ErrUnauthorized,
			want: "401",
		},
		{
			name: "forbidden",
			err:  external.ErrForbidden,
			want: "403",
		},
		{
			name: "not found",
			err:  external.ErrNotFound,
			want: "404",
		},
		{
			name: "rate limited",
			err:  external.ErrRateLimited,
			want: "429",
		},
		{
			name: "timeout",
			err:  external.ErrTimeout,
			want: "timeout",
		},
		{
			name: "upstream",
			err:  external.ErrUpstream,
			want: "5xx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redditStatusFromError(tt.err); got != tt.want {
				t.Fatalf(
					"expected %q, got %q",
					tt.want,
					got,
				)
			}
		})
	}
}

func TestRedditWrappedErrorsRemainClassifiable(
	t *testing.T,
) {
	err := errors.New("wrapped")

	err = errors.Join(
		external.ErrRateLimited,
		err,
	)

	if redditStatusFromError(err) != "429" {
		t.Fatal(
			"expected wrapped rate-limit error to classify as 429",
		)
	}
}

func assertMetricFamilyPresent(
	t *testing.T,
	families []*dto.MetricFamily,
	name string,
) {
	t.Helper()

	for _, family := range families {
		if family.GetName() == name {
			return
		}
	}

	t.Fatalf(
		"metric family %q not found",
		name,
	)
}

func TestRedditObservabilityLogDoesNotContainSensitiveData(
	t *testing.T,
) {
	var logs bytes.Buffer

	logger := slog.New(
		slog.NewTextHandler(
			&logs,
			nil,
		),
	)

	observeOperation(
		logger,
		"authenticate",
		"",
		"200",
		true,
		time.Now(),
		nil,
	)

	output := logs.String()

	for _, forbidden := range []string{
		"client_secret",
		"access_token",
		"Authorization",
		"password",
		"token=",
	} {
		if strings.Contains(
			output,
			forbidden,
		) {
			t.Fatalf(
				"sensitive value leaked into log: %q",
				forbidden,
			)
		}
	}
}
