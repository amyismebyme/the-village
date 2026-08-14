package middleware

import (
	"net/http"
	"strings"
)

func routeLabel(r *http.Request) string {
	// Go 1.22 ServeMux pattern
	if r.Pattern != "" {

		parts := strings.SplitN(
			r.Pattern,
			" ",
			2,
		)

		if len(parts) == 2 {
			return parts[1]
		}

		return r.Pattern
	}

	return "unknown"
}
