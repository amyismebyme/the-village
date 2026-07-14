package server

import "time"

type Runtime struct {
	StartedAt time.Time
	Version   string
}

var AppRuntime = Runtime{
	StartedAt: time.Now(),
	Version:   "0.1.1",
}

func Uptime() time.Duration {
	return time.Since(AppRuntime.StartedAt)
}