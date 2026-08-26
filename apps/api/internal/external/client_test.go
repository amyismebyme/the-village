package external

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientUsesCallerContext(t *testing.T) {
	client := NewClient(
		roundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				if err := req.Context().Err(); err == nil {
					t.Fatal(
						"expected canceled context",
					)
				}

				return nil, req.Context().Err()
			},
		),
		time.Second,
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cancel()

	req := httptest.NewRequest(
		http.MethodGet,
		"https://example.test",
		nil,
	)

	_, err := client.Do(
		ctx,
		req,
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

type roundTripperFunc func(
	*http.Request,
) (*http.Response, error)

func (f roundTripperFunc) Do(
	req *http.Request,
) (*http.Response, error) {
	return f(req)
}

func TestClassifyStatus(t *testing.T) {
	tests := []struct {
		status int
		want   error
	}{
		{
			status: http.StatusUnauthorized,
			want:   ErrUnauthorized,
		},
		{
			status: http.StatusForbidden,
			want:   ErrForbidden,
		},
		{
			status: http.StatusNotFound,
			want:   ErrNotFound,
		},
		{
			status: http.StatusTooManyRequests,
			want:   ErrRateLimited,
		},
		{
			status: http.StatusInternalServerError,
			want:   ErrUpstream,
		},
	}

	for _, tt := range tests {
		t.Run(
			http.StatusText(tt.status),
			func(t *testing.T) {
				if !errors.Is(
					classifyStatus(tt.status),
					tt.want,
				) {
					t.Fatalf(
						"expected %v for status %d",
						tt.want,
						tt.status,
					)
				}
			},
		)
	}
}

func TestClientUsesRequestTimeout(t *testing.T) {
	client := NewClient(
		roundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()

				return nil, req.Context().Err()
			},
		),
		25*time.Millisecond,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"https://example.test",
		nil,
	)

	_, err := client.Do(
		context.Background(),
		req,
	)

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf(
			"expected ErrTimeout, got %v",
			err,
		)
	}
}

func TestClientParentDeadlineWins(t *testing.T) {
	client := NewClient(
		roundTripperFunc(
			func(req *http.Request) (*http.Response, error) {
				<-req.Context().Done()

				return nil, req.Context().Err()
			},
		),
		time.Second,
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		25*time.Millisecond,
	)
	defer cancel()

	req := httptest.NewRequest(
		http.MethodGet,
		"https://example.test",
		nil,
	)

	_, err := client.Do(
		ctx,
		req,
	)

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf(
			"expected ErrTimeout, got %v",
			err,
		)
	}
}
