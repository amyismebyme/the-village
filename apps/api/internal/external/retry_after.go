package external

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ParseRetryAfter(
	value string,
	now time.Time,
) time.Duration {
	value = strings.TrimSpace(value)

	if value == "" {
		return 0
	}

	if seconds, err := strconv.ParseInt(
		value,
		10,
		64,
	); err == nil {
		if seconds <= 0 {
			return 0
		}

		return time.Duration(seconds) * time.Second
	}

	when, err := http.ParseTime(value)
	if err != nil {
		return 0
	}

	delay := time.Until(when)

	if now != nilTime {
		delay = when.Sub(now)
	}

	if delay <= 0 {
		return 0
	}

	return delay
}

var nilTime time.Time
