package version

import "testing"

func TestCurrentReturnsBuildInfo(t *testing.T) {
	originalValue := Value
	originalCommit := Commit
	originalBuildDate := BuildDate
	t.Cleanup(func() {
		Value = originalValue
		Commit = originalCommit
		BuildDate = originalBuildDate
	})

	Value = "1.0.0"
	Commit = "abc123"
	BuildDate = "2026-03-25"

	info := Current()
	if info.Version != "1.0.0" || info.Commit != "abc123" || info.BuildDate != "2026-03-25" {
		t.Fatalf("unexpected build info %#v", info)
	}
}
