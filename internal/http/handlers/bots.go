package handlers

import (
	"net/http"

	"clawbot-server/internal/platform/bots"

	"github.com/go-chi/chi/v5"
)

type BotsHandler struct {
	service bots.Service
}

func NewBotsHandler(service bots.Service) *BotsHandler {
	return &BotsHandler{service: service}
}

func (h *BotsHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *BotsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input bots.CreateInput
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

func (h *BotsHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "botID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *BotsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input bots.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.Update(r.Context(), chi.URLParam(r, "botID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}
