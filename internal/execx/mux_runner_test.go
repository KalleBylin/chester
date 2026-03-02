package execx

import (
	"context"
	"testing"
)

func TestMuxRunnerRoutesByCommandName(t *testing.T) {
	t.Parallel()

	fallback := NewMockRunner(
		Expectation{
			Name: "git",
			Args: []string{"status"},
			Result: Result{Stdout: []byte("git")},
		},
	)
	ghRunner := NewMockRunner(
		Expectation{
			Name: "gh",
			Args: []string{"auth", "status"},
			Result: Result{Stdout: []byte("gh")},
		},
	)

	runner := NewMuxRunner(fallback)
	runner.Set("gh", ghRunner)

	result, err := runner.Run(context.Background(), "gh", "auth", "status")
	if err != nil {
		t.Fatalf("Run() gh error = %v", err)
	}
	if string(result.Stdout) != "gh" {
		t.Fatalf("Run() gh stdout = %q", string(result.Stdout))
	}

	result, err = runner.Run(context.Background(), "git", "status")
	if err != nil {
		t.Fatalf("Run() git error = %v", err)
	}
	if string(result.Stdout) != "git" {
		t.Fatalf("Run() git stdout = %q", string(result.Stdout))
	}
}
