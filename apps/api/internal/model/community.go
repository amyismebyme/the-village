package model

import "time"

type Community struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	URL            string    `json:"url"`
	ExternalSource string    `json:"external_source"`
	Category       string    `json:"category"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
