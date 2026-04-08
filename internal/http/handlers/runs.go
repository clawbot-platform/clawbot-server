package handlers

import (
	"net/http"
	"strings"

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

func (h *RunsHandler) StartRun(w http.ResponseWriter, r *http.Request) {
	var input runs.ExecuteRunInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.StartRun(r.Context(), chi.URLParam(r, "runID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) ExecuteCycleRun(w http.ResponseWriter, r *http.Request) {
	var input runs.ExecuteRunInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.ExecuteCycleRun(
		r.Context(),
		chi.URLParam(r, "runID"),
		chi.URLParam(r, "cycleID"),
		input,
		actorFromRequest(r),
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) AttachArtifact(w http.ResponseWriter, r *http.Request) {
	var input runs.AttachArtifactInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.AttachArtifact(r.Context(), chi.URLParam(r, "runID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"data": item})
}

func (h *RunsHandler) ListArtifacts(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListArtifacts(r.Context(), chi.URLParam(r, "runID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": items})
}

func (h *RunsHandler) CreateCycle(w http.ResponseWriter, r *http.Request) {
	var input runs.CreateCycleInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.CreateCycle(r.Context(), chi.URLParam(r, "runID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"data": item})
}

func (h *RunsHandler) UpdateCycle(w http.ResponseWriter, r *http.Request) {
	var input runs.UpdateCycleInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.UpdateCycle(r.Context(), chi.URLParam(r, "runID"), chi.URLParam(r, "cycleID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) GetCycle(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetCycle(r.Context(), chi.URLParam(r, "runID"), chi.URLParam(r, "cycleID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) UpsertComparison(w http.ResponseWriter, r *http.Request) {
	var input runs.UpsertComparisonInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.UpsertComparison(r.Context(), chi.URLParam(r, "runID"), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) GetComparison(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetComparison(r.Context(), chi.URLParam(r, "runID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) RegisterModelProfile(w http.ResponseWriter, r *http.Request) {
	var input runs.RegisterModelProfileInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	item, err := h.service.RegisterModelProfile(r.Context(), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, envelope{"data": item})
}

func (h *RunsHandler) GetModelProfile(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetModelProfile(r.Context(), chi.URLParam(r, "modelProfileID"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

func (h *RunsHandler) DependencyHealth(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.DependencyHealth(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}

type reviewActionBody struct {
	ReviewerID       string  `json:"reviewer_id"`
	ReviewerType     string  `json:"reviewer_type"`
	Rationale        string  `json:"rationale"`
	CycleID          *string `json:"cycle_id,omitempty"`
	PolicyDecisionID *string `json:"policy_decision_id,omitempty"`
}

func (h *RunsHandler) ApproveRun(w http.ResponseWriter, r *http.Request) {
	h.handleReviewAction(w, r, "approve")
}

func (h *RunsHandler) RejectRun(w http.ResponseWriter, r *http.Request) {
	h.handleReviewAction(w, r, "reject")
}

func (h *RunsHandler) OverrideRun(w http.ResponseWriter, r *http.Request) {
	h.handleReviewAction(w, r, "override")
}

func (h *RunsHandler) DeferRun(w http.ResponseWriter, r *http.Request) {
	h.handleReviewAction(w, r, "defer")
}

func (h *RunsHandler) handleReviewAction(w http.ResponseWriter, r *http.Request, action string) {
	var body reviewActionBody
	if r.Body != nil {
		_ = decodeJSON(r, &body)
	}

	item, err := h.service.ReviewAction(r.Context(), chi.URLParam(r, "runID"), runs.ReviewActionInput{
		Action:           strings.TrimSpace(action),
		ReviewerID:       strings.TrimSpace(body.ReviewerID),
		ReviewerType:     strings.TrimSpace(body.ReviewerType),
		Rationale:        strings.TrimSpace(body.Rationale),
		CycleID:          body.CycleID,
		PolicyDecisionID: body.PolicyDecisionID,
	}, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{"data": item})
}
