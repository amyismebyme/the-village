package repository

import "errors"

var (

	// Generic repository errors.

	ErrNotFound = errors.New("resource not found")

	ErrAlreadyExists = errors.New("resource already exists")

	ErrConflict = errors.New("resource conflict")

	ErrInvalidInput = errors.New("invalid repository input")

	ErrNotImplemented = errors.New("repository not implemented")
)
