package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func RepoRoot(tb testing.TB) string {
	tb.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("runtime.Caller(0) failed")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	info, err := os.Stat(root)
	if err != nil {
		tb.Fatalf("stat repo root: %v", err)
	}
	if !info.IsDir() {
		tb.Fatalf("repo root is not a directory: %s", root)
	}
	return root
}

func FixturePath(tb testing.TB, elements ...string) string {
	tb.Helper()

	parts := []string{RepoRoot(tb), "testdata", "fixtures"}
	parts = append(parts, elements...)
	return filepath.Join(parts...)
}

func ReadFixture(tb testing.TB, elements ...string) []byte {
	tb.Helper()

	path := FixturePath(tb, elements...)
	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read fixture %s: %v", path, err)
	}
	return data
}

func SandboxScriptPath(tb testing.TB) string {
	tb.Helper()
	return filepath.Join(RepoRoot(tb), "testdata", "setup_sandbox.sh")
}
