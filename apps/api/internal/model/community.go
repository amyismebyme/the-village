package model

import (
	"strings"
	"time"
)

type Community struct {
	ID             int64
	Name           string
	Slug           string
	Description    string
	ExternalSource string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (c *Community) Normalize() {

	c.Name = strings.TrimSpace(c.Name)

	c.Slug = strings.ToLower(
		strings.TrimSpace(c.Slug),
	)

	c.Description = strings.TrimSpace(
		c.Description,
	)

	c.ExternalSource = strings.TrimSpace(
		c.ExternalSource,
	)
}
