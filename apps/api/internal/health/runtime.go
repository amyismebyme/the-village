package health

import "context"

type RuntimeChecker struct{}

func (RuntimeChecker) Name() string {
	return "runtime"
}

func (RuntimeChecker) Check(ctx context.Context) error {
	return nil
}
