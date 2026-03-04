package cmd

import "testing"

func TestParseLineTargetSupportsInlineLocation(t *testing.T) {
	t.Parallel()

	file, start, end, err := parseLineTarget("db/queries.go:112:115", "")
	if err != nil {
		t.Fatalf("parseLineTarget() error = %v", err)
	}
	if file != "db/queries.go" || start != 112 || end != 115 {
		t.Fatalf("parseLineTarget() = (%q, %d, %d)", file, start, end)
	}
}

func TestParseLineTargetSupportsDashInlineLocation(t *testing.T) {
	t.Parallel()

	file, start, end, err := parseLineTarget("db/queries.go:112-115", "")
	if err != nil {
		t.Fatalf("parseLineTarget() error = %v", err)
	}
	if file != "db/queries.go" || start != 112 || end != 115 {
		t.Fatalf("parseLineTarget() = (%q, %d, %d)", file, start, end)
	}
}

func TestParseLineTargetKeepsLegacyFlag(t *testing.T) {
	t.Parallel()

	file, start, end, err := parseLineTarget("db/queries.go", "112,115")
	if err != nil {
		t.Fatalf("parseLineTarget() error = %v", err)
	}
	if file != "db/queries.go" || start != 112 || end != 115 {
		t.Fatalf("parseLineTarget() = (%q, %d, %d)", file, start, end)
	}
}

func TestParseLineTargetRejectsMixedForms(t *testing.T) {
	t.Parallel()

	_, _, _, err := parseLineTarget("db/queries.go:112:115", "112,115")
	if err == nil {
		t.Fatal("parseLineTarget() error = nil, want mixed-form error")
	}
}
