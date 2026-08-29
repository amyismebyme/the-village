package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
)

const (
	defaultBaseURL = "https://oauth.reddit.com"
)

type Client struct {
	client    *external.Client
	baseURL   *url.URL
	userAgent string
	logger    *slog.Logger

	rateLimiter external.RateLimiter
	retryPolicy *external.RetryPolicy
}

func NewClient(
	httpClient external.HTTPClient,
	baseURL string,
	userAgent string,
	requestTimeout time.Duration,
) (*Client, error) {
	if httpClient == nil {
		return nil, errors.New(
			"reddit client: HTTP client is required",
		)
	}

	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	parsedURL, err := httputil.ParseBaseURL(
		baseURL,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit client: %w",
			err,
		)
	}

	if strings.TrimSpace(userAgent) == "" {
		return nil, errors.New(
			"reddit client: user agent is required",
		)
	}

	return &Client{
		client: external.NewClient(
			httpClient,
			requestTimeout,
		),
		baseURL:   parsedURL,
		userAgent: userAgent,
	}, nil
}

func (c *Client) SetLogger(
	logger *slog.Logger,
) {
	c.logger = logger
}

func (c *Client) SetRateLimiter(
	limiter external.RateLimiter,
) {
	c.rateLimiter = limiter
}

func (c *Client) SetRetryPolicy(
	policy *external.RetryPolicy,
) {
	c.retryPolicy = policy
}

func (c *Client) NewRequest(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
) (*http.Request, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New(
			"reddit client: path is required",
		)
	}

	relativePath := strings.TrimPrefix(
		path,
		"/",
	)

	requestURL := *c.baseURL

	requestURL.Path =
		strings.TrimRight(
			requestURL.Path,
			"/",
		) +
			"/" +
			relativePath

	if query != nil {
		requestURL.RawQuery =
			query.Encode()
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		requestURL.String(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit client: create request: %w",
			err,
		)
	}

	req.Header.Set(
		"User-Agent",
		c.userAgent,
	)

	return req, nil
}

func (c *Client) Do(
	ctx context.Context,
	req *http.Request,
) (*http.Response, error) {
	if req == nil {
		return nil, external.ErrNilRequest
	}

	return c.client.Do(
		ctx,
		req,
	)
}

func (c *Client) DoAuthenticated(
	ctx context.Context,
	req *http.Request,
	accessToken string,
) (*http.Response, error) {
	if req == nil {
		return nil, external.ErrNilRequest
	}

	if strings.TrimSpace(accessToken) == "" {
		return nil, external.ErrUnauthorized
	}

	req = req.Clone(ctx)

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	return c.client.DoChecked(
		ctx,
		req,
	)
}

func DecodeJSON[T any](
	response *http.Response,
	target *T,
) error {
	if response == nil {
		return external.ErrInvalidPayload
	}

	if target == nil {
		return external.ErrInvalidPayload
	}

	body, err := external.ReadAndClose(
		response.Body,
	)
	if err != nil {
		return fmt.Errorf(
			"reddit client: read response: %w",
			err,
		)
	}

	if len(body) == 0 {
		return external.ErrInvalidPayload
	}

	if err := json.Unmarshal(
		body,
		target,
	); err != nil {
		return fmt.Errorf(
			"%w: %v",
			external.ErrInvalidPayload,
			err,
		)
	}

	return nil
}

func (c *Client) FetchListing(
	ctx context.Context,
	accessToken string,
	subreddit string,
	limit int,
	after string,
) (
	listing ListingResponse,
	err error,
) {
	start := time.Now()

	status := "error"
	externalID := ""
	requestAttempted := false

	defer func() {
		observeOperation(
			c.logger,
			"fetch",
			externalID,
			status,
			requestAttempted,
			start,
			err,
		)
	}()

	if err := ctx.Err(); err != nil {
		return ListingResponse{}, err
	}

	if limit <= 0 {
		limit = 25
	}

	query := url.Values{}

	query.Set(
		"limit",
		fmt.Sprintf("%d", limit),
	)

	if after != "" {
		query.Set(
			"after",
			after,
		)
	}

	operation := func(
		operationCtx context.Context,
	) error {
		req, requestErr := c.NewRequest(
			operationCtx,
			http.MethodGet,
			"/r/"+url.PathEscape(
				subreddit,
			)+"/new",
			query,
		)
		if requestErr != nil {
			return requestErr
		}

		if c.rateLimiter != nil {
			if requestErr := c.rateLimiter.Wait(
				operationCtx,
			); requestErr != nil {
				return requestErr
			}
		}

		requestAttempted = true

		response, requestErr := c.DoAuthenticated(
			operationCtx,
			req,
			accessToken,
		)
		if requestErr != nil {
			status = redditStatusFromError(
				requestErr,
			)

			return requestErr
		}

		status = fmt.Sprintf(
			"%d",
			response.StatusCode,
		)

		if requestErr := DecodeJSON(
			response,
			&listing,
		); requestErr != nil {
			status = "invalid_payload"

			return fmt.Errorf(
				"reddit fetch listing: decode response: %w",
				requestErr,
			)
		}

		return nil
	}

	if c.retryPolicy != nil {
		err = c.retryPolicy.Do(
			ctx,
			operation,
			func(event external.RetryEvent) {
				observeRetry(
					c.logger,
					"fetch",
					event,
				)
			},
		)

		if err != nil {
			return ListingResponse{}, fmt.Errorf(
				"reddit fetch listing: %w",
				err,
			)
		}

		return listing, nil
	}

	err = operation(ctx)

	if err != nil {
		return ListingResponse{}, fmt.Errorf(
			"reddit fetch listing: %w",
			err,
		)
	}

	return listing, nil
}
