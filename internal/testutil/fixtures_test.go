package testutil

import (
	"path/filepath"
	"testing"
)

func TestFixtureHelpersUseStablePaths(t *testing.T) {
	t.Parallel()

	root := RepoRoot(t)
	if filepath.Base(root) != "chester" {
		t.Fatalf("RepoRoot() = %q, want repo named chester", root)
	}

	path := FixturePath(t, "gh", "pr_view_123.json")
	if filepath.Base(path) != "pr_view_123.json" {
		t.Fatalf("FixturePath() = %q", path)
	}

	data := ReadFixture(t, "gh", "pr_view_123.json")
	if len(data) == 0 {
		t.Fatal("ReadFixture() returned empty data")
	}
}
