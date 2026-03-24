package handlers

import (
	"net/http"

	"clawbot-server/internal/platform/policies"

	"github.com/go-chi/chi/v5"
)

type PoliciesHandler struct {
	service policies.Service
}

func NewPoliciesHandler(service policies.Service) *PoliciesHandler {
	return &PoliciesHandler{service: service}
}

func (h *PoliciesHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *PoliciesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input policies.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.Create(r.Context(), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"data": item})
}

func (h *PoliciesHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "policyID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *PoliciesHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input policies.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.Update(r.Context(), chi.URLParam(r, "policyID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}
