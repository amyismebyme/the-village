package handlers

import (
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type communityResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description,omitempty"`
	ExternalSource string    `json:"external_source,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newCommunityResponse(
	c *model.Community,
) communityResponse {
	return communityResponse{
		ID:             c.ID,
		Name:           c.Name,
		Slug:           c.Slug,
		Description:    c.Description,
		ExternalSource: c.ExternalSource,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func newCommunityResponses(
	communities []*model.Community,
) []communityResponse {
	if communities == nil {
		return []communityResponse{}
	}

	responses := make([]communityResponse, 0, len(communities))

	for _, community := range communities {
		if community == nil {
			continue
		}

		responses = append(
			responses,
			newCommunityResponse(community),
		)
	}

	return responses
}
