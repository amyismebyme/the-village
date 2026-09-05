package reddit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
)

const (
	accessTokenPath = "/api/v1/access_token"

	grantTypeClientCredentials = "client_credentials"

	contentTypeForm = "application/x-www-form-urlencoded"

	authorizationHeader = "Authorization"
	userAgentHeader     = "User-Agent"
)

type Authenticator struct {
	clientID     string
	clientSecret string
	userAgent    string

	client  *external.Client
	baseURL *url.URL
	logger  *slog.Logger

	rateLimiter external.RateLimiter
	retryPolicy *external.RetryPolicy

	mu    sync.Mutex
	token tokenResponse
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`

	ExpiresAt time.Time `json:"-"`
}

func NewAuthenticator(
	httpClient external.HTTPClient,
	baseURL string,
	clientID string,
	clientSecret string,
	userAgent string,
	requestTimeout time.Duration,
) (*Authenticator, error) {
	if httpClient == nil {
		return nil, errors.New(
			"reddit authenticator: HTTP client is required",
		)
	}

	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New(
			"reddit authenticator: base URL is required",
		)
	}

	if strings.TrimSpace(clientID) == "" {
		return nil, errors.New(
			"reddit authenticator: client ID is required",
		)
	}

	if strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New(
			"reddit authenticator: client secret is required",
		)
	}

	if strings.TrimSpace(userAgent) == "" {
		return nil, errors.New(
			"reddit authenticator: user agent is required",
		)
	}

	parsedURL, err := httputil.ParseBaseURL(
		baseURL,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit authenticator: %w",
			err,
		)
	}

	return &Authenticator{
		clientID:     clientID,
		clientSecret: clientSecret,
		userAgent:    userAgent,

		client: external.NewClient(
			httpClient,
			requestTimeout,
		),

		baseURL: parsedURL,
	}, nil
}

func (a *Authenticator) SetLogger(
	logger *slog.Logger,
) {
	a.logger = logger
}

func (a *Authenticator) SetRateLimiter(
	limiter external.RateLimiter,
) {
	a.rateLimiter = limiter
}

func (a *Authenticator) SetRetryPolicy(
	policy *external.RetryPolicy,
) {
	a.retryPolicy = policy
}

// Token returns a cached application access token when valid.
// Access tokens and client secrets are never logged.
func (a *Authenticator) Token(
	ctx context.Context,
) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return "", err
	}

	if a.token.valid() {
		return a.token.AccessToken, nil
	}

	token, err := a.fetchToken(ctx)
	if err != nil {
		return "", err
	}

	a.token = token

	return token.AccessToken, nil
}

func (a *Authenticator) newTokenRequest(
	ctx context.Context,
) (*http.Request, error) {
	requestURL := *a.baseURL

	requestURL.Path =
		strings.TrimRight(
			requestURL.Path,
			"/",
		) +
			accessTokenPath

	form := url.Values{}

	form.Set(
		"grant_type",
		grantTypeClientCredentials,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL.String(),
		strings.NewReader(
			form.Encode(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit authenticator: create token request: %w",
			err,
		)
	}

	req.Header.Set(
		userAgentHeader,
		a.userAgent,
	)

	req.Header.Set(
		"Content-Type",
		contentTypeForm,
	)

	req.SetBasicAuth(
		a.clientID,
		a.clientSecret,
	)

	return req, nil
}

func (a *Authenticator) fetchToken(
	ctx context.Context,
) (
	result tokenResponse,
	err error,
) {
	start := time.Now()

	status := "error"
	requestAttempted := false

	defer func() {
		observeOperation(
			a.logger,
			"authenticate",
			"",
			status,
			requestAttempted,
			start,
			err,
		)
	}()

	operation := func(
		operationCtx context.Context,
	) error {
		req, requestErr := a.newTokenRequest(
			operationCtx,
		)
		if requestErr != nil {
			return requestErr
		}

		if a.rateLimiter != nil {
			if requestErr := a.rateLimiter.Wait(
				operationCtx,
			); requestErr != nil {
				return requestErr
			}
		}

		requestAttempted = true
		attemptStart := time.Now()

		response, requestErr := a.client.DoChecked(
			operationCtx,
			req,
		)
		if requestErr != nil {
			status = redditStatusFromError(
				requestErr,
			)

			observeRequestAttempt(
				"authenticate",
				status,
				attemptStart,
				requestErr,
			)

			return requestErr
		}

		status = fmt.Sprintf(
			"%d",
			response.StatusCode,
		)

		var token tokenResponse

		requestPayloadErr := func() error {
			if decodeErr := DecodeJSON(
				response,
				&token,
			); decodeErr != nil {
				return fmt.Errorf(
					"%w: decode token response",
					external.ErrInvalidPayload,
				)
			}

			if strings.TrimSpace(
				token.AccessToken,
			) == "" {
				return fmt.Errorf(
					"%w: missing access token",
					external.ErrInvalidPayload,
				)
			}

			if strings.TrimSpace(
				token.TokenType,
			) == "" {
				return fmt.Errorf(
					"%w: missing token type",
					external.ErrInvalidPayload,
				)
			}

			if token.ExpiresIn <= 0 {
				return fmt.Errorf(
					"%w: invalid token expiry",
					external.ErrInvalidPayload,
				)
			}

			return nil
		}()

		if requestPayloadErr != nil {
			status = "invalid_payload"

			observeRequestAttempt(
				"authenticate",
				fmt.Sprintf("%d", response.StatusCode),
				attemptStart,
				requestPayloadErr,
			)

			return requestPayloadErr
		}

		observeRequestAttempt(
			"authenticate",
			status,
			attemptStart,
			nil,
		)

		token.ExpiresAt = time.Now().Add(
			time.Duration(
				token.ExpiresIn,
			) * time.Second,
		)

		result = token

		return nil
	}

	if a.retryPolicy != nil {
		err = a.retryPolicy.Do(
			ctx,
			operation,
			func(event external.RetryEvent) {
				observeRetry(
					a.logger,
					"authenticate",
					event,
				)
			},
		)

		if err != nil {
			return tokenResponse{}, fmt.Errorf(
				"reddit authenticator: token request: %w",
				err,
			)
		}

		return result, nil
	}

	err = operation(ctx)

	if err != nil {
		return tokenResponse{}, fmt.Errorf(
			"reddit authenticator: token request: %w",
			err,
		)
	}

	return result, nil
}

func (t tokenResponse) valid() bool {
	if strings.TrimSpace(
		t.AccessToken,
	) == "" {
		return false
	}

	if t.ExpiresAt.IsZero() {
		return false
	}

	return time.Now().Add(
		30 * time.Second,
	).Before(
		t.ExpiresAt,
	)
}
