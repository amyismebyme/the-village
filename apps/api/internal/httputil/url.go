package httputil

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func ParseBaseURL(
	value string,
	allowHTTP bool,
) (*url.URL, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New(
			"base URL is required",
		)
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf(
			"parse base URL: %w",
			err,
		)
	}

	if parsed.Host == "" {
		return nil, errors.New(
			"base URL must include a host",
		)
	}

	if allowHTTP {
		if parsed.Scheme != "http" &&
			parsed.Scheme != "https" {
			return nil, errors.New(
				"base URL must use http or https",
			)
		}
	} else if parsed.Scheme != "https" {
		return nil, errors.New(
			"base URL must use https",
		)
	}

	return parsed, nil
}
