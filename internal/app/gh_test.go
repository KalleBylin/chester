package app

import (
	"context"
	"errors"
	"testing"

	"chester/internal/execx"
	"chester/internal/testutil"
)

func TestRunGHReturnsStdout(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "rate_limit"},
			Result: execx.Result{
				Stdout: []byte(`{"ok":true}`),
			},
		},
	)

	got, err := RunGH(context.Background(), runner, "api", "rate_limit")
	if err != nil {
		t.Fatalf("RunGH() error = %v", err)
	}
	if string(got) != `{"ok":true}` {
		t.Fatalf("RunGH() = %q", string(got))
	}
}

func TestRunGHMapsMissingBinary(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{"auth", "status"},
			Err:  errors.Join(execx.ErrCommandNotFound, errors.New("gh")),
		},
	)

	_, err := RunGH(context.Background(), runner, "auth", "status")
	if !errors.Is(err, ErrGHNotInstalled) {
		t.Fatalf("RunGH() error = %v, want ErrGHNotInstalled", err)
	}
}

func TestRunGHMapsAuthenticationFailures(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "user"},
			Err: &execx.RunError{
				Name:     "gh",
				Args:     []string{"api", "user"},
				ExitCode: 4,
				Stderr:   []byte("You are not logged into any GitHub hosts. Run gh auth login to authenticate."),
			},
		},
	)

	_, err := RunGH(context.Background(), runner, "api", "user")
	if !errors.Is(err, ErrGHUnauthenticated) {
		t.Fatalf("RunGH() error = %v, want ErrGHUnauthenticated", err)
	}
}

func TestGHWrappersBuildExactCommands(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{"pr", "view", "123", "--repo", "acme/chester", "--json", "number,title,body"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pr_view_123.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"issue", "view", "456", "--repo", "acme/chester", "--json", "number,title,body"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "issue_view_456.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "repos/acme/chester/commits/deadbeef/pulls"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "commit_pulls_sha_with_pr.json")},
		},
	)

	if _, err := GHPRView(context.Background(), runner, "acme/chester", "123", []string{"number", "title", "body"}); err != nil {
		t.Fatalf("GHPRView() error = %v", err)
	}
	if _, err := GHIssueView(context.Background(), runner, "acme/chester", "456", []string{"number", "title", "body"}); err != nil {
		t.Fatalf("GHIssueView() error = %v", err)
	}
	if _, err := GHAPI(context.Background(), runner, "repos/acme/chester/commits/deadbeef/pulls"); err != nil {
		t.Fatalf("GHAPI() error = %v", err)
	}
	if err := runner.AssertDone(); err != nil {
		t.Fatalf("AssertDone() error = %v", err)
	}
}
