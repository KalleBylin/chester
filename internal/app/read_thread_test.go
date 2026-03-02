package app

import (
	"context"
	"strings"
	"testing"

	"chester/internal/execx"
	"chester/internal/testutil"
)

func TestReadThreadRendersPRConversation(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"pr", "view", "123", "--repo", "acme/chester",
				"--json", "number,title,body,url,state,isDraft,comments,reviews",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "pr_view_123_with_reviews.json"),
			},
		},
	)

	got, err := ReadThread(context.Background(), runner, "acme/chester", "123")
	if err != nil {
		t.Fatalf("ReadThread() error = %v", err)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "read_thread_pr.md")))
	if strings.TrimSpace(got) != want {
		t.Fatalf("ReadThread() mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestReadThreadFallsBackToIssue(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"pr", "view", "456", "--repo", "acme/chester",
				"--json", "number,title,body,url,state,isDraft,comments,reviews",
			},
			Err: &execx.RunError{
				Name:     "gh",
				Args:     []string{"pr", "view", "456"},
				ExitCode: 1,
				Stderr:   []byte("no pull request found for #456"),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"issue", "view", "456", "--repo", "acme/chester",
				"--json", "number,title,body,url,state,comments",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "issue_view_456.json"),
			},
		},
	)

	got, err := ReadThread(context.Background(), runner, "acme/chester", "456")
	if err != nil {
		t.Fatalf("ReadThread() error = %v", err)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "read_thread_issue.md")))
	if strings.TrimSpace(got) != want {
		t.Fatalf("ReadThread() mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestReadThreadUsesEmptyBodyPlaceholder(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"pr", "view", "999", "--repo", "acme/chester",
				"--json", "number,title,body,url,state,isDraft,comments,reviews",
			},
			Result: execx.Result{
				Stdout: []byte(`{"number":999,"title":"Empty","body":"<!-- hidden -->","url":"https://github.com/acme/chester/pull/999","state":"MERGED","isDraft":false,"comments":[],"reviews":[]}`),
			},
		},
	)

	got, err := ReadThread(context.Background(), runner, "acme/chester", "999")
	if err != nil {
		t.Fatalf("ReadThread() error = %v", err)
	}

	if !strings.Contains(got, "## Body\n(empty)") {
		t.Fatalf("ReadThread() = %q, want empty body placeholder", got)
	}
	if !strings.Contains(got, "## Thread\n(none)") {
		t.Fatalf("ReadThread() = %q, want empty thread placeholder", got)
	}
}
