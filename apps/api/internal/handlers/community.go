package handlers

import (
	"encoding/json"
	"net/http"

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
//	POST /communities
func (h *Handler) CreateCommunity(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		w.Header().Set(
			"Allow",
			http.MethodPost,
		)

		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method_not_allowed",
			"method not allowed",
		)

		return
	}

	defer r.Body.Close()

	var request createCommunityRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"request body contains invalid JSON",
		)

		return
	}

	// Reject multiple JSON values in the request body.
	var extra any

	if err := decoder.Decode(&extra); err == nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"request body must contain exactly one JSON object",
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
		writeCommunityServiceError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		community,
	)
}
