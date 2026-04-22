package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"clawbot-server/internal/watchlistidentity"
)

type IdentityIntegrationService interface {
	CompareRecords(
		ctx context.Context,
		tenantID string,
		caseID string,
		correlationID string,
		input watchlistidentity.CompareRecordsInput,
	) (watchlistidentity.CompareRecordsResult, error)

	ScreenSubjectAgainstOFAC(
		ctx context.Context,
		tenantID string,
		caseID string,
		correlationID string,
		subject watchlistidentity.Subject,
	) (watchlistidentity.OFACScreeningResult, error)
}

type IdentityIntegrationHandler struct {
	service IdentityIntegrationService
}

func NewIdentityIntegrationHandler(service IdentityIntegrationService) *IdentityIntegrationHandler {
	return &IdentityIntegrationHandler{service: service}
}

func (h *IdentityIntegrationHandler) Compare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "identity integration not configured"})
		return
	}

	var req struct {
		TenantID string                                `json:"tenant_id"`
		Input    watchlistidentity.CompareRecordsInput `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	correlationID := correlationIDFromRequest(r, "")
	caseID := strings.TrimSpace(r.Header.Get("X-Case-ID"))
	w.Header().Set("X-Correlation-ID", correlationID)

	resp, err := h.service.CompareRecords(r.Context(), req.TenantID, caseID, correlationID, req.Input)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *IdentityIntegrationHandler) ScreenOFAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if h.service == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "identity integration not configured"})
		return
	}

	var req struct {
		TenantID string                    `json:"tenant_id"`
		CaseID   string                    `json:"case_id"`
		Subject  watchlistidentity.Subject `json:"subject"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	correlationID := correlationIDFromRequest(r, "")
	w.Header().Set("X-Correlation-ID", correlationID)

	resp, err := h.service.ScreenSubjectAgainstOFAC(r.Context(), req.TenantID, req.CaseID, correlationID, req.Subject)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func correlationIDFromRequest(r *http.Request, fallback string) string {
	if v := strings.TrimSpace(r.Header.Get("X-Correlation-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(fallback); v != "" {
		return v
	}
	return "corr_" + time.Now().UTC().Format("20060102150405.000000000")
}

// writeJSON is defined in common.go.
