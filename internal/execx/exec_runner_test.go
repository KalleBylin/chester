package execx

import (
	"context"
	"errors"
	"testing"
)

func TestExecRunnerRunsCommand(t *testing.T) {
	t.Parallel()

	runner := ExecRunner{}
	result, err := runner.Run(context.Background(), "sh", "-c", "printf hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if string(result.Stdout) != "hello" {
		t.Fatalf("Run() stdout = %q", string(result.Stdout))
	}
}

func TestExecRunnerReturnsCommandNotFound(t *testing.T) {
	t.Parallel()

	runner := ExecRunner{}
	_, err := runner.Run(context.Background(), "chester-command-does-not-exist")
	if !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("Run() error = %v, want ErrCommandNotFound", err)
	}
}

func TestExecRunnerReturnsRunError(t *testing.T) {
	t.Parallel()

	runner := ExecRunner{}
	result, err := runner.Run(context.Background(), "sh", "-c", "printf no >&2; exit 7")
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}

	runErr := &RunError{}
	if !errors.As(err, &runErr) {
		t.Fatalf("Run() error = %T, want *RunError", err)
	}
	if runErr.ExitCode != 7 {
		t.Fatalf("RunError.ExitCode = %d, want 7", runErr.ExitCode)
	}
	if string(result.Stderr) != "no" {
		t.Fatalf("Run() stderr = %q, want %q", string(result.Stderr), "no")
	}
}
