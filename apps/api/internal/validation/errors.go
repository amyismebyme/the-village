package validation

import "fmt"

// FieldError associates a validation failure with the
// domain field that caused it while preserving the original
// validation error through Unwrap().
type FieldError struct {
	Field string
	Err   error
}

func NewFieldError(
	field string,
	err error,
) error {
	if err == nil {
		return nil
	}

	return FieldError{
		Field: field,
		Err:   err,
	}
}

func (e FieldError) Error() string {
	if e.Err == nil {
		return e.Field
	}

	return fmt.Sprintf(
		"%s: %v",
		e.Field,
		e.Err,
	)
}

func (e FieldError) Unwrap() error {
	return e.Err
}
