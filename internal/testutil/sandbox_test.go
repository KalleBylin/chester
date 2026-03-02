package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSandboxScriptBuildsDeterministicRepoShape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	repoDir := filepath.Join(root, "sandbox")

	cmd := exec.Command(SandboxScriptPath(t), repoDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("setup_sandbox.sh failed: %v\n%s", err, string(output))
	}

	manifestPath := filepath.Join(repoDir, ".chester-sandbox-manifest")
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
	if len(lines) < 8 {
		t.Fatalf("manifest lines = %d, want at least 8", len(lines))
	}

	wantKeys := []string{
		"SESSION_FILE=",
		"DB_FILE=",
		"RANGE_FROM=",
		"RANGE_TO=",
		"BLAME_START=",
		"BLAME_END=",
		"SQUASH_SHA=",
		"MERGE_SHA=",
		"DIRECT_SHA=",
	}
	for _, key := range wantKeys {
		if !strings.Contains(string(manifest), key) {
			t.Fatalf("manifest missing %q\n%s", key, string(manifest))
		}
	}

	if _, err := os.Stat(filepath.Join(repoDir, "internal", "auth", "session.go")); err != nil {
		t.Fatalf("session.go missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoDir, "db", "queries.go")); err != nil {
		t.Fatalf("queries.go missing: %v", err)
	}
}
