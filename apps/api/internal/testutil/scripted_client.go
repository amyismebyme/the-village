package testutil

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
)

// HTTPOutcome describes one deterministic HTTP result returned by a
// ScriptedHTTPClient. When Err is non-nil, the response is not returned.
type HTTPOutcome struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	Err        error
}

// ScriptedHTTPClient is a provider-neutral failure-injection harness for
// tests. Each configured path consumes outcomes in order and falls back to
// the wrapped client after the scripted outcomes are exhausted.
type ScriptedHTTPClient struct {
	Base HTTPClient

	mu sync.Mutex

	scripts map[string][]HTTPOutcome
	index   map[string]int
	calls   map[string]int
}

// HTTPClient is the minimal outbound HTTP contract needed by the harness.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewScriptedHTTPClient(
	base HTTPClient,
) *ScriptedHTTPClient {
	if base == nil {
		panic(
			"scripted HTTP client: base client is required",
		)
	}

	return &ScriptedHTTPClient{
		Base:    base,
		scripts: make(map[string][]HTTPOutcome),
		index:   make(map[string]int),
		calls:   make(map[string]int),
	}
}

// SetScript replaces the scripted outcomes for one request path.
// Outcomes are consumed in order, one outcome per matching request.
func (c *ScriptedHTTPClient) SetScript(
	path string,
	outcomes ...HTTPOutcome,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.scripts[path] =
		append(
			[]HTTPOutcome(nil),
			outcomes...,
		)

	c.index[path] = 0
}

// ResetScript resets a configured path to its first scripted outcome.
func (c *ScriptedHTTPClient) ResetScript(
	path string,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.index[path] = 0
}

// Calls returns the number of requests observed for the given path.
func (c *ScriptedHTTPClient) Calls(
	path string,
) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.calls[path]
}

func (c *ScriptedHTTPClient) Do(
	req *http.Request,
) (*http.Response, error) {
	if req == nil {
		return nil, errors.New(
			"scripted HTTP client: request is required",
		)
	}

	path := req.URL.Path

	c.mu.Lock()

	c.calls[path]++

	outcomes, scripted := c.scripts[path]
	position := c.index[path]

	if scripted &&
		position < len(outcomes) {
		outcome := outcomes[position]
		c.index[path] = position + 1

		c.mu.Unlock()

		if outcome.Err != nil {
			return nil, outcome.Err
		}

		return outcomeResponse(
			req,
			outcome,
		), nil
	}

	c.mu.Unlock()

	return c.Base.Do(req)
}

func outcomeResponse(
	req *http.Request,
	outcome HTTPOutcome,
) *http.Response {
	headers := make(
		http.Header,
		len(outcome.Headers),
	)

	for key, value := range outcome.Headers {
		headers.Set(
			key,
			value,
		)
	}

	body := io.NopCloser(
		strings.NewReader(outcome.Body),
	)

	statusCode := outcome.StatusCode

	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	return &http.Response{
		StatusCode: statusCode,
		Header:     headers,
		Body:       body,
		Request:    req,
	}
}
