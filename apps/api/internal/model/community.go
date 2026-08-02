package model

import "time"

// Community represents a Village community.
type Community struct {
	ID int64

	Name string

	Slug string

	Category string

	Description string

	ExternalSource string

	CreatedAt time.Time

	UpdatedAt time.Time
}
