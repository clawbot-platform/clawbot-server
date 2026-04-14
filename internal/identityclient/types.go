package identityclient

type RecordRef struct {
	SourceSystem   string `json:"source_system"`
	SourceRecordID string `json:"source_record_id"`
}

type CompareRequest struct {
	TenantID string    `json:"tenant_id"`
	Left     RecordRef `json:"left"`
	Right    RecordRef `json:"right"`
	Explain  bool      `json:"explain"`
}

type CompareSourceRef struct {
	SourceSystem   string `json:"source_system"`
	SourceRecordID string `json:"source_record_id"`
}

type CompareExplanation struct {
	ExplanationID string             `json:"explanation_id"`
	Summary       string             `json:"summary"`
	Why           []string           `json:"why"`
	WhyNot        []string           `json:"why_not"`
	How           []string           `json:"how"`
	SourceRefs    []CompareSourceRef `json:"source_refs"`
}

type CompareResponse struct {
	Disposition     string             `json:"disposition"`
	ConfidenceBand  string             `json:"confidence_band"`
	Explanation     CompareExplanation `json:"explanation"`
	DecisionTraceID string             `json:"decision_trace_id"`
}

type OFACSubject struct {
	Name        string            `json:"name"`
	DOB         string            `json:"dob,omitempty"`
	Country     string            `json:"country,omitempty"`
	Identifiers map[string]string `json:"identifiers,omitempty"`
}

type ScreenOFACRequest struct {
	TenantID string      `json:"tenant_id"`
	CaseID   string      `json:"case_id"`
	Subject  OFACSubject `json:"subject"`
}

type OFACCandidate struct {
	DatasetRunID string `json:"dataset_run_id"`
	ListKind     string `json:"list_kind"`
	ListUID      string `json:"list_uid"`
	Name         string `json:"name"`
	MatchedOn    string `json:"matched_on"`
	Score        int    `json:"score"`
	NeedsReview  bool   `json:"needs_review"`
}

type ScreenOFACResponse struct {
	ScreeningID     string          `json:"screening_id"`
	Decision        string          `json:"decision"`
	DecisionTraceID string          `json:"decision_trace_id"`
	Candidates      []OFACCandidate `json:"candidates"`
}
