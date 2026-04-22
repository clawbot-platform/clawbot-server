package watchlistreviewclient

import "encoding/json"

type ReviewOptions struct {
	Explain bool   `json:"explain,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

type ReviewRequest struct {
	TenantID     string          `json:"tenant_id"`
	CaseID       string          `json:"case_id,omitempty"`
	SourceSystem string          `json:"source_system"`
	RawAlert     json.RawMessage `json:"raw_alert"`
	Options      ReviewOptions   `json:"options,omitempty"`
}

type IdentityTraceRefs struct {
	DecisionTraceID string `json:"decision_trace_id,omitempty"`
	ExplanationID   string `json:"explanation_id,omitempty"`
	ScreeningID     string `json:"screening_id,omitempty"`
}

type ReviewResponse struct {
	Status            string            `json:"status"`
	CaseID            string            `json:"case_id,omitempty"`
	AlertID           string            `json:"alert_id,omitempty"`
	Warnings          []string          `json:"warnings,omitempty"`
	IdentityTraceRefs IdentityTraceRefs `json:"identity_trace_refs,omitempty"`
	ReviewContext     any               `json:"review_context,omitempty"`
}
