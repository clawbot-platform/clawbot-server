package store

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queryRowStub struct {
	values []int64
	err    error
}

func (s queryRowStub) Scan(dest ...any) error {
	if s.err != nil {
		return s.err
	}
	for i := range dest {
		ptr, ok := dest[i].(*int64)
		if !ok {
			return errors.New("unexpected destination type")
		}
		*ptr = s.values[i]
	}
	return nil
}

type dbtxStub struct {
	row queryRowStub
}

func (dbtxStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (dbtxStub) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (s dbtxStub) QueryRow(context.Context, string, ...any) pgx.Row {
	return s.row
}

func TestDashboardReaderSummary(t *testing.T) {
	reader := NewDashboardReader(dbtxStub{
		row: queryRowStub{values: []int64{1, 2, 3, 4}},
	})

	summary, err := reader.Summary(context.Background())
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.Runs != 1 || summary.AuditEvents != 4 {
		t.Fatalf("unexpected summary %#v", summary)
	}
}

func TestDashboardReaderSummaryError(t *testing.T) {
	reader := NewDashboardReader(dbtxStub{
		row: queryRowStub{err: errors.New("boom")},
	})

	_, err := reader.Summary(context.Background())
	if err == nil {
		t.Fatal("expected summary error")
	}
}
