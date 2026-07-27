package repository

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrInvalidID      = errors.New("invalid id")
	ErrNotImplemented = errors.New("repository method not implemented")
)
