package service

import "errors"

var (
	ErrCommunityAlreadyExists = errors.New("community already exists")
	ErrInvalidCommunity       = errors.New("invalid community")
	ErrInvalidCommunityID     = errors.New("invalid community id")
	ErrNilCommunity           = errors.New("community is required")
	ErrInvalidPagination      = errors.New("invalid pagination")
)
