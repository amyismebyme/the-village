package health

import "context"

type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

type Result struct {
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}
