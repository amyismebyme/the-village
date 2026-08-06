/*
Package api provides reusable HTTP helpers used by all API handlers.

Responsibilities:

  - JSON encoding
  - Standard response envelope
  - Consistent error responses
  - Shared HTTP utilities

Every HTTP handler should use this package instead of directly
calling json.NewEncoder or http.Error.

Example:

	api.WriteJSON(
	    w,
	    http.StatusOK,
	    community,
	)

	api.NotFound(
	    w,
	    "community not found",
	)
*/
package api
