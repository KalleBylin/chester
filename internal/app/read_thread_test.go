package app

import (
	"context"
	"strings"
	"testing"

	"github.com/KalleBylin/chester/internal/execx"
	"github.com/KalleBylin/chester/internal/testutil"
)

func TestReadThreadRendersPRConversation(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"pr", "view", "123", "--repo", "acme/chester",
				"--json", "number,title,body,url,state,isDraft",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "pr_view_123_metadata.json"),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/issues/123/comments?per_page=100",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "issue_comments_123.json"),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/pulls/123/reviews?per_page=100",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "pull_reviews_123.json"),
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
				"--json", "number,title,body,url,state,isDraft",
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
				"--json", "number,title,body,url,state",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "issue_view_456_metadata.json"),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/issues/456/comments?per_page=100",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "issue_comments_456.json"),
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
				"--json", "number,title,body,url,state,isDraft",
			},
			Result: execx.Result{
				Stdout: []byte(`{"number":999,"title":"Empty","body":"<!-- hidden -->","url":"https://github.com/acme/chester/pull/999","state":"MERGED","isDraft":false}`),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/issues/999/comments?per_page=100",
			},
			Result: execx.Result{
				Stdout: []byte(`[]`),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/pulls/999/reviews?per_page=100",
			},
			Result: execx.Result{
				Stdout: []byte(`[]`),
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
