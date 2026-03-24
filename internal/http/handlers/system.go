package handlers

import (
	"context"
	"net/http"

	"clawbot-server/internal/version"
)

type SystemHandler struct {
	readiness func(context.Context) error
	buildInfo version.Info
}

func NewSystemHandler(readiness func(context.Context) error, buildInfo version.Info) *SystemHandler {
	return &SystemHandler{readiness: readiness, buildInfo: buildInfo}
}

func (h *SystemHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{"status": "ok"})
}

func (h *SystemHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.readiness(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "not_ready", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, envelope{"status": "ready"})
}

func (h *SystemHandler) Version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.buildInfo)
}
