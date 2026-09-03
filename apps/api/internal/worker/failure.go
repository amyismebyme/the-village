package worker

import (
	"context"
	"errors"
)

type FailurePolicy int

const (
	FailurePolicyContinue FailurePolicy = iota
	FailurePolicyStop
)

func ClassifyFailure(
	err error,
) FailurePolicy {
	if err == nil {
		return FailurePolicyContinue
	}

	if errors.Is(
		err,
		context.Canceled,
	) {
		return FailurePolicyStop
	}

	return FailurePolicyContinue
}
