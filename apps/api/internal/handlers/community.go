package handlers

import (
	"net/http"
	"strconv"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
)

type createCommunityRequest struct {
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
	ExternalSource string `json:"external_source"`
}

// CreateCommunity handles:
//
//	POST /api/v1/communities
func (h *Handler) CreateCommunity(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)

		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"method not allowed",
		)

		return
	}

	var request createCommunityRequest

	if err := decodeJSON(w, r, &request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			err.Error(),
		)

		return
	}

	community := &model.Community{
		Name:           request.Name,
		Slug:           request.Slug,
		Description:    request.Description,
		ExternalSource: request.ExternalSource,
	}

	if err := h.communityService.Create(
		r.Context(),
		community,
	); err != nil {
		writeServiceError(w, err)
		return
	}

	httputil.WriteJSON(
		w,
		http.StatusCreated,
		newCommunityResponse(community),
	)
}

// ListCommunities handles:
//
//	GET /api/v1/communities
func (h *Handler) ListCommunities(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)

		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"method not allowed",
		)

		return
	}

	communities, err := h.communityService.List(
		r.Context(),
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if communities == nil {
		communities = []*model.Community{}
	}

	response := struct {
		Communities []communityResponse `json:"communities"`
	}{
		Communities: newCommunityResponses(communities),
	}

	httputil.WriteJSON(
		w,
		http.StatusOK,
		response,
	)
}

// GetCommunity handles:
//
//	GET /api/v1/communities/{id}
func (h *Handler) GetCommunity(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)

		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"method not allowed",
		)

		return
	}

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"invalid community id",
		)

		return
	}

	community, err := h.communityService.Get(
		r.Context(),
		id,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if community == nil {
		writeError(
			w,
			http.StatusNotFound,
			"community_not_found",
			"community not found",
		)

		return
	}

	httputil.WriteJSON(
		w,
		http.StatusOK,
		newCommunityResponse(community),
	)
}

// UpdateCommunity handles:
//
//	PUT /api/v1/communities/{id}
//
// UpdateCommunity handles:
//
//	PUT /api/v1/communities/{id}
func (h *Handler) UpdateCommunity(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)

		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"method not allowed",
		)

		return
	}

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"invalid community id",
		)

		return
	}

	var request createCommunityRequest

	if err := decodeJSON(
		w,
		r,
		&request,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_request",
			err.Error(),
		)

		return
	}

	community := &model.Community{
		ID:             id,
		Name:           request.Name,
		Slug:           request.Slug,
		Description:    request.Description,
		ExternalSource: request.ExternalSource,
	}

	if err := h.communityService.Update(
		r.Context(),
		community,
	); err != nil {
		writeServiceError(w, err)
		return
	}

	httputil.WriteJSON(
		w,
		http.StatusOK,
		newCommunityResponse(community),
	)
}

// DeleteCommunity handles:
//
//	DELETE /api/v1/communities/{id}
func (h *Handler) DeleteCommunity(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)

		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"method not allowed",
		)

		return
	}

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"invalid community id",
		)

		return
	}

	if err := h.communityService.Delete(
		r.Context(),
		id,
	); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
