package handlers

import "github.com/amyismebyme/the-village/apps/api/internal/service"

type Handler struct {
	communityService service.CommunityService
}

func NewHandler(
	communityService service.CommunityService,
) *Handler {

	return &Handler{
		communityService: communityService,
	}
}
