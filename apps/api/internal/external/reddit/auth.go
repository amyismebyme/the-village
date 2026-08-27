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
	"sync"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/external"
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

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf(
			"reddit authenticator: parse base URL: %w",
			err,
		)
	}

	if parsedURL.Scheme != "http" &&
		parsedURL.Scheme != "https" {
		return nil, errors.New(
			"reddit authenticator: base URL must use http or https",
		)
	}

	if parsedURL.Host == "" {
		return nil, errors.New(
			"reddit authenticator: base URL must include host",
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

// Token returns a cached application token when it remains valid.
// Tokens are kept in memory only.
func (a *Authenticator) Token(
	ctx context.Context,
) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

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

func (a *Authenticator) fetchToken(
	ctx context.Context,
) (result tokenResponse, err error) {
	start := time.Now()
	status := "error"

	defer func() {
		observeOperation(
			a.logger,
			"authenticate",
			"",
			status,
			start,
			err,
		)
	}()

	requestURL := *a.baseURL

	requestURL.Path = strings.TrimRight(
		requestURL.Path,
		"/",
	) + accessTokenPath

	form := url.Values{}
	form.Set(
		"grant_type",
		grantTypeClientCredentials,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL.String(),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return tokenResponse{}, fmt.Errorf(
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

	response, err := a.client.DoChecked(
		ctx,
		req,
	)
	if err != nil {
		status = redditStatusFromError(err)

		return tokenResponse{}, fmt.Errorf(
			"reddit authenticator: token request: %w",
			err,
		)
	}

	status = fmt.Sprintf(
		"%d",
		response.StatusCode,
	)

	defer response.Body.Close()

	var token tokenResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&token); err != nil {
		return tokenResponse{}, fmt.Errorf(
			"%w: decode token response",
			external.ErrInvalidPayload,
		)
	}

	if strings.TrimSpace(token.AccessToken) == "" {
		return tokenResponse{}, fmt.Errorf(
			"%w: missing access token",
			external.ErrInvalidPayload,
		)
	}

	if token.TokenType == "" {
		return tokenResponse{}, fmt.Errorf(
			"%w: missing token type",
			external.ErrInvalidPayload,
		)
	}

	if token.ExpiresIn <= 0 {
		return tokenResponse{}, fmt.Errorf(
			"%w: invalid token expiry",
			external.ErrInvalidPayload,
		)
	}

	token.ExpiresAt = time.Now().Add(
		time.Duration(token.ExpiresIn) * time.Second,
	)

	return token, nil
}

func (t tokenResponse) valid() bool {
	if strings.TrimSpace(t.AccessToken) == "" {
		return false
	}

	if t.ExpiresAt.IsZero() {
		return false
	}

	return time.Now().Add(
		30 * time.Second,
	).Before(t.ExpiresAt)
}
