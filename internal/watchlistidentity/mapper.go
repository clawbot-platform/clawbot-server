package watchlistidentity

import "clawbot-server/internal/identityclient"

func mapCompareResponse(response identityclient.CompareResponse) CompareRecordsResult {
	return CompareRecordsResult{
		Disposition:    response.Disposition,
		ConfidenceBand: response.ConfidenceBand,
		Explanation: CompareExplanation{
			ExplanationID: response.Explanation.ExplanationID,
			Summary:       response.Explanation.Summary,
			Why:           append([]string(nil), response.Explanation.Why...),
			WhyNot:        append([]string(nil), response.Explanation.WhyNot...),
			How:           append([]string(nil), response.Explanation.How...),
			SourceRefs:    mapCompareSourceRefs(response.Explanation.SourceRefs),
		},
		DecisionTraceID: response.DecisionTraceID,
	}
}

func mapCompareSourceRefs(sourceRefs []identityclient.CompareSourceRef) []CompareSourceRef {
	if len(sourceRefs) == 0 {
		return nil
	}

	mapped := make([]CompareSourceRef, 0, len(sourceRefs))
	for _, sourceRef := range sourceRefs {
		mapped = append(mapped, CompareSourceRef{
			SourceSystem:   sourceRef.SourceSystem,
			SourceRecordID: sourceRef.SourceRecordID,
		})
	}

	return mapped
}

func mapOFACScreeningResponse(response identityclient.ScreenOFACResponse) OFACScreeningResult {
	return OFACScreeningResult{
		ScreeningID:     response.ScreeningID,
		Decision:        response.Decision,
		DecisionTraceID: response.DecisionTraceID,
		Candidates:      mapOFACCandidates(response.Candidates),
	}
}

func mapOFACCandidates(candidates []identityclient.OFACCandidate) []OFACCandidate {
	if len(candidates) == 0 {
		return nil
	}

	mapped := make([]OFACCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		mapped = append(mapped, OFACCandidate{
			DatasetRunID: candidate.DatasetRunID,
			ListKind:     candidate.ListKind,
			ListUID:      candidate.ListUID,
			Name:         candidate.Name,
			MatchedOn:    candidate.MatchedOn,
			Score:        candidate.Score,
			NeedsReview:  candidate.NeedsReview,
		})
	}

	return mapped
}
