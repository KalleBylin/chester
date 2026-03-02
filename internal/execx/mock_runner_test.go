package execx

import (
	"context"
	"errors"
	"testing"
)

func TestMockRunnerMatchesExactCommand(t *testing.T) {
	t.Parallel()

	runner := NewMockRunner(
		Expectation{
			Name: "gh",
			Args: []string{"pr", "view", "123"},
			Result: Result{
				Stdout: []byte("ok"),
			},
		},
	)

	result, err := runner.Run(context.Background(), "gh", "pr", "view", "123")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if string(result.Stdout) != "ok" {
		t.Fatalf("Run() stdout = %q", string(result.Stdout))
	}
	if err := runner.AssertDone(); err != nil {
		t.Fatalf("AssertDone() error = %v", err)
	}
}

func TestMockRunnerRejectsUnexpectedCommand(t *testing.T) {
	t.Parallel()

	runner := NewMockRunner(
		Expectation{Name: "git", Args: []string{"status"}},
	)

	_, err := runner.Run(context.Background(), "git", "log")
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestMockRunnerReportsUnconsumedExpectation(t *testing.T) {
	t.Parallel()

	runner := NewMockRunner(
		Expectation{Name: "git", Args: []string{"status"}},
	)

	if err := runner.AssertDone(); err == nil {
		t.Fatal("AssertDone() error = nil, want error")
	}
}

func TestMockRunnerReturnsConfiguredError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	runner := NewMockRunner(
		Expectation{
			Name: "gh",
			Args: []string{"auth", "status"},
			Err:  wantErr,
		},
	)

	_, err := runner.Run(context.Background(), "gh", "auth", "status")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}
