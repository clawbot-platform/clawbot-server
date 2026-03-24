package handlers

import (
	"context"
	"net/http"

	"clawbot-server/internal/platform/store"
)

type DashboardService interface {
	Summary(context.Context) (store.DashboardSummary, error)
}

type DashboardHandler struct {
	service DashboardService
}

func NewDashboardHandler(service DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.Summary(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": summary})
}
