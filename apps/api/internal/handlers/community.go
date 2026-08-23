package handlers

import (
	"net/http"
	"strconv"
	"time"

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

	if r.Method != http.MethodPost {
		status = http.StatusMethodNotAllowed

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

	limit := service.DefaultCommunityPageLimit
	offset := 0

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid_pagination",
				"limit must be a positive integer",
			)
			return
		}

		limit = parsed
	}

	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid_pagination",
				"offset must be a non-negative integer",
			)
			return
		}

		offset = parsed
	}

	page, err := h.communityService.List(
		r.Context(),
		limit,
		offset,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response := struct {
		Communities []communityResponse `json:"communities"`
		Pagination  struct {
			Limit  int   `json:"limit"`
			Offset int   `json:"offset"`
			Total  int64 `json:"total"`
		} `json:"pagination"`
	}{
		Communities: newCommunityResponses(page.Communities),
	}

	response.Pagination.Limit = page.Limit
	response.Pagination.Offset = page.Offset
	response.Pagination.Total = page.Total

	httpStatus := http.StatusOK
	httputil.WriteJSON(
		w,
		httpStatus,
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

	if r.Method != http.MethodPut {
		status = http.StatusMethodNotAllowed

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

	if r.Method != http.MethodDelete {
		status = http.StatusMethodNotAllowed

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
