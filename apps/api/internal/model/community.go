package model

import (
	"strings"
	"time"
)

type Community struct {
	ID int64 `json:"id"`

	Name string `json:"name"`

	Slug string `json:"slug"`

	Description string `json:"description,omitempty"`

	ExternalSource string `json:"external_source,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
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
