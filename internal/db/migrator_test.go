package db

import "testing"

func TestLoadMigrationsFindsSortedPairs(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected embedded migrations")
	}
	for i, migration := range migrations {
		if migration.upFile == "" || migration.downFile == "" {
			t.Fatalf("expected migration pair, got %#v", migration)
		}
		if i > 0 && migrations[i-1].version >= migration.version {
			t.Fatalf("expected sorted migrations, got %#v", migrations)
		}
	}
}
