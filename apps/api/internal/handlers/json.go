package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxJSONBodySize = 1 << 20 // 1 MiB

func decodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) error {
	contentType := r.Header.Get("Content-Type")

	if contentType == "" {
		return fmt.Errorf("content type is required")
	}

	// Accept application/json and application/json;charset=UTF-8.
	if contentType != "application/json" &&
		contentType != "application/json; charset=utf-8" &&
		contentType != "application/json;charset=utf-8" {
		return fmt.Errorf("content type must be application/json")
	}

	if r.Body == nil {
		return fmt.Errorf("request body is required")
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxJSONBodySize,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("request body is required")
		}

		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return fmt.Errorf("request body is too large")
		}

		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Reject a second JSON value.
	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("request body must contain exactly one JSON object")
		}

		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}
