package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://oauth.reddit.com"
)

type Client struct {
	client    *external.Client
	baseURL   *url.URL
	userAgent string
	logger    *slog.Logger
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

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit client: parse base URL: %w",
			err,
		)
	}

	if parsedURL.Scheme != "http" &&
		parsedURL.Scheme != "https" {
		return nil, errors.New(
			"reddit client: base URL must use http or https",
		)
	}

	if parsedURL.Host == "" {
		return nil, errors.New(
			"reddit client: base URL must include host",
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

	relativePath := strings.TrimPrefix(path, "/")

	requestURL := *c.baseURL
	requestURL.Path = strings.TrimRight(
		requestURL.Path,
		"/",
	) + "/" + relativePath

	if query != nil {
		requestURL.RawQuery = query.Encode()
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
		userAgentHeader,
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
		authorizationHeader,
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

	if err := json.Unmarshal(body, target); err != nil {
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
) (listing ListingResponse, err error) {
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

	if strings.TrimSpace(subreddit) == "" {
		err = fmt.Errorf(
			"%w: subreddit is required",
			external.ErrInvalidConfig,
		)

		return ListingResponse{}, err
	}

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

	req, err := c.NewRequest(
		ctx,
		http.MethodGet,
		"/r/"+url.PathEscape(subreddit)+"/new",
		query,
	)
	if err != nil {
		return ListingResponse{}, err
	}

	requestAttempted = true

	response, err := c.DoAuthenticated(
		ctx,
		req,
		accessToken,
	)
	if err != nil {
		status = redditStatusFromError(err)

		return ListingResponse{}, fmt.Errorf(
			"reddit fetch listing: %w",
			err,
		)
	}

	status = fmt.Sprintf(
		"%d",
		response.StatusCode,
	)

	if err := DecodeJSON(
		response,
		&listing,
	); err != nil {
		return ListingResponse{}, fmt.Errorf(
			"reddit fetch listing: decode response: %w",
			err,
		)
	}

	return listing, nil
}
