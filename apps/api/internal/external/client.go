package external

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is the minimal outbound HTTP contract used by
// external integrations.
//
// The interface intentionally matches the useful subset of
// net/http.Client so production code can use *http.Client while
// tests can provide a deterministic fake.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client provides common behavior for an external integration.
type Client struct {
	HTTP           HTTPClient
	RequestTimeout time.Duration
}

// NewClient creates a provider-neutral external HTTP client.
func NewClient(
	httpClient HTTPClient,
	requestTimeout time.Duration,
) *Client {
	if httpClient == nil {
		panic("external client: HTTP client is required")
	}

	return &Client{
		HTTP:           httpClient,
		RequestTimeout: requestTimeout,
	}
}

// Do executes an outbound HTTP request using the caller's context.
//
// The caller owns the context and therefore controls cancellation
// and the lifetime of the external request.
func (c *Client) Do(
	ctx context.Context,
	req *http.Request,
) (*http.Response, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	requestContext, cancel := WithRequestTimeout(
		ctx,
		c.RequestTimeout,
	)
	defer cancel()

	req = req.WithContext(requestContext)

	response, err := c.HTTP.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf(
				"%w: %w",
				ErrTimeout,
				err,
			)
		}

		if errors.Is(err, context.Canceled) {
			return nil, err
		}

		return nil, fmt.Errorf(
			"%w: %v",
			ErrUpstream,
			err,
		)
	}

	return response, nil
}

// ReadAndClose reads a response body and always closes it.
func ReadAndClose(
	body io.ReadCloser,
) ([]byte, error) {
	if body == nil {
		return nil, ErrNilResponseBody
	}

	data, readErr := io.ReadAll(body)
	closeErr := body.Close()

	if readErr != nil {
		return nil, readErr
	}

	if closeErr != nil {
		return nil, closeErr
	}

	return data, nil
}


func classifyStatus(status int) error {
	switch {
	case status >= 200 && status < 300:
		return nil

	case status == http.StatusUnauthorized:
		return ErrUnauthorized

	case status == http.StatusForbidden:
		return ErrForbidden

	case status == http.StatusNotFound:
		return ErrNotFound

	case status == http.StatusTooManyRequests:
		return ErrRateLimited

	case status >= 500:
		return ErrUpstream

	default:
		return ErrInvalidPayload
	}
}

func (c *Client) DoChecked(
	ctx context.Context,
	req *http.Request,
) (*http.Response, error) {
	response, err := c.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := classifyStatus(response.StatusCode); err != nil {
		_ = response.Body.Close()

		return nil, fmt.Errorf(
			"%w: status=%d",
			err,
			response.StatusCode,
		)
	}

	return response, nil
}
