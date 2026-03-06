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
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/pulls/123/comments?per_page=100",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "pull_comments_123.json"),
			},
		},
	)

	got, err := ReadThread(context.Background(), runner, "acme/chester", "123")
	if err != nil {
		t.Fatalf("ReadThread() error = %v", err)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "read_thread_pr.md")))
	rendered := strings.TrimSpace(RenderReadThreadMarkdown(got))
	if rendered != want {
		t.Fatalf("ReadThread() mismatch\nwant:\n%s\n\ngot:\n%s", want, rendered)
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
	rendered := strings.TrimSpace(RenderReadThreadMarkdown(got))
	if rendered != want {
		t.Fatalf("ReadThread() mismatch\nwant:\n%s\n\ngot:\n%s", want, rendered)
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
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/pulls/999/comments?per_page=100",
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

	rendered := RenderReadThreadMarkdown(got)
	if !strings.Contains(rendered, "## Body\n(empty)") {
		t.Fatalf("ReadThread() = %q, want empty body placeholder", rendered)
	}
	if !strings.Contains(rendered, "## Comments\n(none)") {
		t.Fatalf("ReadThread() = %q, want empty comments placeholder", rendered)
	}
	if !strings.Contains(rendered, "## Reviews\n(none)") {
		t.Fatalf("ReadThread() = %q, want empty reviews placeholder", rendered)
	}
	if !strings.Contains(rendered, "## Review Comments\n(none)") {
		t.Fatalf("ReadThread() = %q, want empty review comments placeholder", rendered)
	}
}
