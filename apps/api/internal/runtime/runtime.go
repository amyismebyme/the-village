package runtime

import (
	"runtime"
	"time"
)

var StartedAt = time.Now()
var BuildVersion = "0.1.2"
// To pull from git later
var GitCommit = "local"
// to pull from os later hardcoded for now
var BuildTime =  time.Now().Add(-10 * time.Second)
var Environment = "dev"

func Uptime() time.Duration {
	return time.Since(StartedAt)
}

func GoVersion() string {
	return runtime.Version()
}
