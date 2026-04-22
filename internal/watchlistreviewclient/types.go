package watchlistreviewclient

import "encoding/json"

type AnalystAlignedReason struct {
	Kind         string   `json:"kind"`
	Strength     string   `json:"strength,omitempty"`
	Message      string   `json:"message"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
}

type AnalystAlignedExplanation struct {
	Summary       string                 `json:"summary"`
	Reasons       []AnalystAlignedReason `json:"reasons,omitempty"`
	AnalystNote   string                 `json:"analyst_note,omitempty"`
	EvidenceKinds []string               `json:"evidence_kinds,omitempty"`
}

type ReviewOptions struct {
	Explain bool   `json:"explain,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

type CompareExplanation struct {
	ExplanationID string                     `json:"explanation_id,omitempty"`
	Summary       string                     `json:"summary,omitempty"`
	Why           []string                   `json:"why,omitempty"`
	WhyNot        []string                   `json:"why_not,omitempty"`
	How           []string                   `json:"how,omitempty"`
	Alignment     *AnalystAlignedExplanation `json:"alignment,omitempty"`
}

type IdentityTraceRefs struct {
	DecisionTraceID string `json:"decision_trace_id,omitempty"`
	ExplanationID   string `json:"explanation_id,omitempty"`
	ScreeningID     string `json:"screening_id,omitempty"`
}

type IdentityCompareContext struct {
	Disposition    string             `json:"disposition,omitempty"`
	ConfidenceBand string             `json:"confidence_band,omitempty"`
	Explanation    CompareExplanation `json:"explanation,omitempty"`
}

type IdentityOFACContext struct {
	Decision string `json:"decision,omitempty"`
}

type IdentityReviewContext struct {
	TraceRefs     IdentityTraceRefs       `json:"trace_refs,omitempty"`
	Compare       *IdentityCompareContext `json:"compare,omitempty"`
	OFACScreening *IdentityOFACContext    `json:"ofac_screening,omitempty"`
}

type ReviewRequest struct {
	TenantID        string                 `json:"tenant_id"`
	CaseID          string                 `json:"case_id,omitempty"`
	SourceSystem    string                 `json:"source_system"`
	RawAlert        json.RawMessage        `json:"raw_alert"`
	Options         ReviewOptions          `json:"options,omitempty"`
	IdentityContext *IdentityReviewContext `json:"identity_context,omitempty"`
}

type ReviewResponse struct {
	Status            string                 `json:"status"`
	CaseID            string                 `json:"case_id,omitempty"`
	AlertID           string                 `json:"alert_id,omitempty"`
	Warnings          []string               `json:"warnings,omitempty"`
	IdentityTraceRefs IdentityTraceRefs      `json:"identity_trace_refs,omitempty"`
	ReviewContext     any                    `json:"review_context,omitempty"`
	IdentityContext   *IdentityReviewContext `json:"identity_context,omitempty"`
}
