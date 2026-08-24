package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/amyismebyme/the-village/apps/api/internal/httputil"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"github.com/amyismebyme/the-village/apps/api/internal/service"
)

type createCommunityRequest struct {
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
	ExternalSource string `json:"external_source"`
}

// updateCommunityRequest intentionally mirrors the fields accepted by PUT
// while keeping create/update contracts distinct at the handler boundary.
type updateCommunityRequest struct {
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
	start := time.Now()

	status := http.StatusInternalServerError
	communityID := int64(0)

	defer func() {
		h.logCommunityOperation(
			r,
			"create",
			communityID,
			status,
			start,
		)
	}()

	var request createCommunityRequest

	if err := decodeJSON(
		w,
		r,
		&request,
	); err != nil {
		status = http.StatusBadRequest

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
		status = writeServiceError(w, err)
		return
	}

	communityID = community.ID
	status = http.StatusCreated

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

	limit := service.DefaultCommunityPageLimit
	offset := 0

	query := r.URL.Query()

	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid_pagination",
				"invalid limit",
			)
			return
		}

		if parsed < 1 {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid_pagination",
				"limit must be greater than 0",
			)
			return
		}

		limit = parsed
	}

	if raw := query.Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid_pagination",
				"invalid offset",
			)
			return
		}

		if parsed < 0 {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid_pagination",
				"offset must be greater than or equal to 0",
			)
			return
		}

		offset = parsed
	}

	result, err := h.communityService.List(
		r.Context(),
		limit,
		offset,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPagination) {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid_pagination",
				"invalid pagination parameters",
			)
			return
		}

		writeServiceError(w, err)
		return
	}

	response := communityListResponse{
		Communities: newCommunityResponses(result.Communities),
		Pagination: paginationResponse{
			Limit:  result.Limit,
			Offset: result.Offset,
			Total:  result.Total,
		},
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
func (h *Handler) UpdateCommunity(
	w http.ResponseWriter,
	r *http.Request,
) {
	start := time.Now()

	status := http.StatusInternalServerError
	communityID := int64(0)

	defer func() {
		h.logCommunityOperation(
			r,
			"update",
			communityID,
			status,
			start,
		)
	}()

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		status = http.StatusBadRequest

		writeError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"invalid community id",
		)

		return
	}

	communityID = id

	var request updateCommunityRequest

	if err := decodeJSON(
		w,
		r,
		&request,
	); err != nil {
		status = http.StatusBadRequest

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
		status = writeServiceError(w, err)
		return
	}

	status = http.StatusOK

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
	start := time.Now()

	status := http.StatusInternalServerError
	communityID := int64(0)

	defer func() {
		h.logCommunityOperation(
			r,
			"delete",
			communityID,
			status,
			start,
		)
	}()

	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || id <= 0 {
		status = http.StatusBadRequest

		writeError(
			w,
			http.StatusBadRequest,
			"invalid_id",
			"invalid community id",
		)

		return
	}

	communityID = id

	if err := h.communityService.Delete(
		r.Context(),
		id,
	); err != nil {
		status = writeServiceError(w, err)
		return
	}

	status = http.StatusNoContent

	w.WriteHeader(http.StatusNoContent)
}
