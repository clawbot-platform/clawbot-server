package watchlistidentity

import (
	"context"
	"fmt"

	"clawbot-server/internal/identityclient"
)

type identityAPI interface {
	Compare(ctx context.Context, req identityclient.CompareRequest, correlationID string, caseID string) (identityclient.CompareResponse, error)
	ScreenOFAC(ctx context.Context, req identityclient.ScreenOFACRequest, correlationID string) (identityclient.ScreenOFACResponse, error)
}

type Service struct {
	identity identityAPI
}

func NewService(identity identityAPI) *Service {
	return &Service{identity: identity}
}

type RecordLocator struct {
	SourceSystem   string `json:"source_system"`
	SourceRecordID string `json:"source_record_id"`
}

type CompareRecordsInput struct {
	Left    RecordLocator `json:"left"`
	Right   RecordLocator `json:"right"`
	Explain bool          `json:"explain"`
}

type CompareSourceRef struct {
	SourceSystem   string `json:"source_system"`
	SourceRecordID string `json:"source_record_id"`
}

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

type CompareExplanation struct {
	ExplanationID string                     `json:"explanation_id"`
	Summary       string                     `json:"summary"`
	Why           []string                   `json:"why"`
	WhyNot        []string                   `json:"why_not"`
	How           []string                   `json:"how"`
	SourceRefs    []CompareSourceRef         `json:"source_refs"`
	Alignment     *AnalystAlignedExplanation `json:"alignment,omitempty"`
}

type CompareRecordsResult struct {
	Disposition     string             `json:"disposition"`
	ConfidenceBand  string             `json:"confidence_band"`
	Explanation     CompareExplanation `json:"explanation"`
	DecisionTraceID string             `json:"decision_trace_id"`
}

type Subject struct {
	Name        string            `json:"name"`
	DOB         string            `json:"dob,omitempty"`
	Country     string            `json:"country,omitempty"`
	Identifiers map[string]string `json:"identifiers,omitempty"`
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

type OFACScreeningResult struct {
	ScreeningID     string          `json:"screening_id"`
	Decision        string          `json:"decision"`
	DecisionTraceID string          `json:"decision_trace_id"`
	Candidates      []OFACCandidate `json:"candidates"`
}

func (s *Service) CompareRecords(
	ctx context.Context,
	tenantID string,
	caseID string,
	correlationID string,
	input CompareRecordsInput,
) (CompareRecordsResult, error) {
	if s.identity == nil {
		return CompareRecordsResult{}, fmt.Errorf("identity client is not configured")
	}

	response, err := s.identity.Compare(ctx, identityclient.CompareRequest{
		TenantID: tenantID,
		Left: identityclient.RecordRef{
			SourceSystem:   input.Left.SourceSystem,
			SourceRecordID: input.Left.SourceRecordID,
		},
		Right: identityclient.RecordRef{
			SourceSystem:   input.Right.SourceSystem,
			SourceRecordID: input.Right.SourceRecordID,
		},
		Explain: input.Explain,
	}, correlationID, caseID)
	if err != nil {
		return CompareRecordsResult{}, fmt.Errorf("compare records: %w", err)
	}

	return mapCompareResponse(response), nil
}

func (s *Service) ScreenSubjectAgainstOFAC(
	ctx context.Context,
	tenantID string,
	caseID string,
	correlationID string,
	subject Subject,
) (OFACScreeningResult, error) {
	if s.identity == nil {
		return OFACScreeningResult{}, fmt.Errorf("identity client is not configured")
	}

	response, err := s.identity.ScreenOFAC(ctx, identityclient.ScreenOFACRequest{
		TenantID: tenantID,
		CaseID:   caseID,
		Subject: identityclient.OFACSubject{
			Name:        subject.Name,
			DOB:         subject.DOB,
			Country:     subject.Country,
			Identifiers: cloneStringMap(subject.Identifiers),
		},
	}, correlationID)
	if err != nil {
		return OFACScreeningResult{}, fmt.Errorf("screen subject against ofac: %w", err)
	}

	return mapOFACScreeningResponse(response), nil
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	copied := make(map[string]string, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}
