package handlers

import "github.com/amyismebyme/the-village/apps/api/internal/service"

type Handler struct {
	communityService service.CommunityService
}

func NewHandler(
	communityService service.CommunityService,
) *Handler {
	if communityService == nil {
		panic("handlers: community service is required")
	}

	return &Handler{
		communityService: communityService,
	}
}
