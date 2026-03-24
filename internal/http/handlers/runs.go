package handlers

import (
	"net/http"

	"clawbot-server/internal/platform/runs"

	"github.com/go-chi/chi/v5"
)

type RunsHandler struct {
	service runs.Service
}

func NewRunsHandler(service runs.Service) *RunsHandler {
	return &RunsHandler{service: service}
}

func (h *RunsHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *RunsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input runs.CreateInput
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

func (h *RunsHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "runID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input runs.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.Update(r.Context(), chi.URLParam(r, "runID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}
