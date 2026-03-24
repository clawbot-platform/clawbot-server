package store

import "context"

type DashboardSummary struct {
	Runs        int64 `json:"runs"`
	Bots        int64 `json:"bots"`
	Policies    int64 `json:"policies"`
	AuditEvents int64 `json:"audit_events"`
}

type DashboardReader struct {
	db DBTX
}

func NewDashboardReader(db DBTX) *DashboardReader {
	return &DashboardReader{db: db}
}

func (r *DashboardReader) Summary(ctx context.Context) (DashboardSummary, error) {
	const query = `
SELECT
  (SELECT COUNT(*) FROM runs) AS runs,
  (SELECT COUNT(*) FROM bots) AS bots,
  (SELECT COUNT(*) FROM policies) AS policies,
  (SELECT COUNT(*) FROM audit_events) AS audit_events
`

	var summary DashboardSummary
	if err := r.db.QueryRow(ctx, query).Scan(
		&summary.Runs,
		&summary.Bots,
		&summary.Policies,
		&summary.AuditEvents,
	); err != nil {
		return DashboardSummary{}, err
	}

	return summary, nil
}
