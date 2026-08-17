package model

import "time"

type Resource struct {
	ID          int64
	Title       string
	Description string
	URL         string
	Category    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
