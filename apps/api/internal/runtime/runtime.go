package runtime

import (
	"runtime"
	"time"
)

var StartedAt = time.Now()

// These variables may be overridden at build time with -ldflags -X.
var BuildVersion = "0.1.2"
var GitCommit = "local"
var BuildTimestamp = ""
var Environment = "dev"

var BuildTime = buildTime()

func buildTime() time.Time {
	if BuildTimestamp != "" {
		if value, err := time.Parse(time.RFC3339, BuildTimestamp); err == nil {
			return value
		}
	}

	return StartedAt
}

func Uptime() time.Duration {
	return time.Since(StartedAt)
}

func GoVersion() string {
	return runtime.Version()
}

func Build() string {
	return BuildVersion
}
