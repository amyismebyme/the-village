package handlers

import (
	"encoding/json"
	"github.com/amyismebyme/the-village/apps/api/internal/model"
	"net/http"
	"strconv"
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

// ListCommunities handles:
//
//	GET /api/v1/communities
func (h *Handler) ListCommunities(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		w.Header().Set(
			"Allow",
			http.MethodGet,
		)

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
		writeCommunityServiceError(w, err)
		return
	}

	// Maintain a stable JSON contract:
	//
	// {
	//     "communities": []
	// }
	//
	// rather than:
	//
	// {
	//     "communities": null
	// }
	if communities == nil {
		communities = []*model.Community{}
	}

	response := struct {
		Communities []*model.Community `json:"communities"`
	}{
		Communities: communities,
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}

func (h *Handler) GetCommunity(w http.ResponseWriter, r *http.Request) {
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

	idString := r.PathValue("id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid_community_id",
			"invalid community id",
		)
		return
	}

	community, err := h.communityService.Get(
		r.Context(),
		id,
	)
	if err != nil {
		writeCommunityServiceError(w, err)
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

	writeJSON(
		w,
		http.StatusOK,
		community,
	)
}

func (h *Handler) UpdateCommunity(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPut {
		writeError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
			"method not allowed",
		)
		return
	}

	idString := r.PathValue("id")

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid community id",
			"invalid community id",
		)
		return
	}

	var community model.Community

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&community); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"malformed JSON",
			"malformed JSON",
		)
		return
	}

	// The resource ID comes from the URL.
	// The client does not control the ID through the request body.
	community.ID = id

	if err := h.communityService.Update(
		r.Context(),
		&community,
	); err != nil {
		writeCommunityServiceError(
			w,
			err,
		)
		return
	}

	// Fetch the updated resource so the API returns the
	// canonical representation, including timestamps/database values.
	updated, err := h.communityService.Get(
		r.Context(),
		id,
	)
	if err != nil {
		writeCommunityServiceError(
			w,
			err,
		)
		return
	}

	if updated == nil {
		writeError(
			w,
			http.StatusNotFound,
			"community not found",
			"community not found",
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		updated,
	)
}
