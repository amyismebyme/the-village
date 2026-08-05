package repository

import "errors"

var (
	ErrNotFound       = errors.New("repository not found")
	ErrAlreadyExists  = errors.New("repository already exists")
	ErrInvalidID      = errors.New("repository invalid id")
	ErrNotImplemented = errors.New("repository method not implemented")
)
