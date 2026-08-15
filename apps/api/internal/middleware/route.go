package middleware

import (
	"net/http"
	"strings"
)

// routeLabel returns the normalized route pattern for logging and metrics.
//
// For Go ServeMux patterns such as:
//
//	GET /api/v1/communities/{id}
//
// it returns:
//
//	/api/v1/communities/{id}
//
// If no ServeMux pattern is available, it falls back to the request path.
func routeLabel(r *http.Request) string {
	if r.Pattern != "" {
		pattern := r.Pattern

		// Go ServeMux may include the HTTP method in the pattern.
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
