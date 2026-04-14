package watchlistidentity

import (
	"context"
	"testing"

	"clawbot-server/internal/identityclient"
)

type stubIdentityClient struct {
	compareResponse identityclient.CompareResponse
	screenResponse  identityclient.ScreenOFACResponse
	lastCompareReq  identityclient.CompareRequest
	lastScreenReq   identityclient.ScreenOFACRequest
}

func (s *stubIdentityClient) Compare(_ context.Context, req identityclient.CompareRequest, _ string, _ string) (identityclient.CompareResponse, error) {
	s.lastCompareReq = req
	return s.compareResponse, nil
}

func (s *stubIdentityClient) ScreenOFAC(_ context.Context, req identityclient.ScreenOFACRequest, _ string) (identityclient.ScreenOFACResponse, error) {
	s.lastScreenReq = req
	return s.screenResponse, nil
}

func TestCompareRecordsMapsResponse(t *testing.T) {
	t.Parallel()

	stub := &stubIdentityClient{
		compareResponse: identityclient.CompareResponse{
			Disposition:    "resolved",
			ConfidenceBand: "high",
			Explanation: identityclient.CompareExplanation{
				ExplanationID: "exp_123",
				Summary:       "match",
				Why:           []string{"same identifier"},
				SourceRefs: []identityclient.CompareSourceRef{
					{SourceSystem: "kyc", SourceRecordID: "left-1"},
				},
			},
			DecisionTraceID: "dt_1",
		},
	}
	service := NewService(stub)

	result, err := service.CompareRecords(context.Background(), "tenant-1", "case-1", "corr-1", CompareRecordsInput{
		Left:    RecordLocator{SourceSystem: "kyc", SourceRecordID: "left-1"},
		Right:   RecordLocator{SourceSystem: "watchlist", SourceRecordID: "right-1"},
		Explain: true,
	})
	if err != nil {
		t.Fatalf("CompareRecords() error = %v", err)
	}

	if stub.lastCompareReq.TenantID != "tenant-1" {
		t.Fatalf("unexpected tenant id %q", stub.lastCompareReq.TenantID)
	}
	if result.DecisionTraceID != "dt_1" {
		t.Fatalf("unexpected decision trace id %q", result.DecisionTraceID)
	}
	if len(result.Explanation.SourceRefs) != 1 {
		t.Fatalf("unexpected source refs %#v", result.Explanation.SourceRefs)
	}
}

func TestScreenSubjectAgainstOFACMapsResponse(t *testing.T) {
	t.Parallel()

	stub := &stubIdentityClient{
		screenResponse: identityclient.ScreenOFACResponse{
			ScreeningID:     "scr_123",
			Decision:        "manual_review",
			DecisionTraceID: "dt_2",
			Candidates: []identityclient.OFACCandidate{
				{
					DatasetRunID: "run-1",
					ListKind:     "sdn",
					ListUID:      "uid-1",
					Name:         "Jane Citizen",
					MatchedOn:    "name",
					Score:        95,
					NeedsReview:  true,
				},
			},
		},
	}
	service := NewService(stub)

	result, err := service.ScreenSubjectAgainstOFAC(context.Background(), "tenant-1", "case-1", "corr-1", Subject{
		Name:        "Jane Citizen",
		Country:     "US",
		Identifiers: map[string]string{"passport": "P1234567"},
	})
	if err != nil {
		t.Fatalf("ScreenSubjectAgainstOFAC() error = %v", err)
	}

	if stub.lastScreenReq.CaseID != "case-1" {
		t.Fatalf("unexpected case id %q", stub.lastScreenReq.CaseID)
	}
	if result.ScreeningID != "scr_123" {
		t.Fatalf("unexpected screening id %q", result.ScreeningID)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("unexpected candidates %#v", result.Candidates)
	}
}
