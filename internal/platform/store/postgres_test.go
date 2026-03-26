package store

import "testing"

func TestNewPostgresExposesPool(t *testing.T) {
	pg := NewPostgres(nil)
	if pg == nil {
		t.Fatal("expected postgres wrapper")
	}
	if pg.Pool() != nil {
		t.Fatalf("expected nil pool, got %#v", pg.Pool())
	}
}
