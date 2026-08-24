package httputil

import (
	"net/http"
	"strings"
)

// RouteLabel returns the normalized route pattern for logs and metrics.
//
// ServeMux method-aware patterns may include the HTTP method, for example
// "GET /api/v1/communities/{id}". This function returns only the route path.
// When no pattern is available it falls back to the request path.
func RouteLabel(r *http.Request) string {
	if r.Pattern != "" {
		pattern := r.Pattern

		if method, route, ok := strings.Cut(pattern, " "); ok {
			switch method {
			case http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodConnect,
				http.MethodOptions,
				http.MethodTrace:
				return route
			}
		}

		return pattern
	}

	if r.URL != nil && r.URL.Path != "" {
		return r.URL.Path
	}

	return "/"
}
