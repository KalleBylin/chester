package app

import (
	"context"
	"strings"
	"testing"

	"github.com/KalleBylin/chester/internal/execx"
	"github.com/KalleBylin/chester/internal/testutil"
)

func TestTextHistoryRendersChronologicalTimeline(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name:   "git",
			Args:   []string{"rev-parse", "--is-inside-work-tree"},
			Result: execx.Result{Stdout: []byte("true\n")},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"log", "--reverse", "--format=%H%x09%s", "-S", "SessionStore", "--", "internal/auth/session.go"},
			Result: execx.Result{
				Stdout: []byte(strings.Join([]string{
					"1111111111111111111111111111111111111111\tExtract session store (#98)",
					"2222222222222222222222222222222222222222\tRename SessionStore helper",
				}, "\n")),
			},
		},
		execx.Expectation{
			Name:   "gh",
			Args:   []string{"pr", "view", "98", "--repo", "acme/chester", "--json", "number,title,body,url,mergedAt"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pr_view_98.json")},
		},
		execx.Expectation{
			Name:   "gh",
			Args:   []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/2222222222222222222222222222222222222222/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s%n%n%b", "2222222222222222222222222222222222222222"},
			Result: execx.Result{
				Stdout: []byte("Rename SessionStore helper\n\nRename SessionStore helper to match new package layout.\n"),
			},
		},
	)

	got, err := TextHistory(context.Background(), runner, "acme/chester", "SessionStore", "internal/auth/session.go")
	if err != nil {
		t.Fatalf("TextHistory() error = %v", err)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "text_history.md")))
	rendered := strings.TrimSpace(RenderTextHistoryMarkdown(got))
	if rendered != want {
		t.Fatalf("TextHistory() mismatch\nwant:\n%s\n\ngot:\n%s", want, rendered)
	}
}

func TestWhyFileWorksWithoutGitHubEnrichment(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name:   "git",
			Args:   []string{"rev-parse", "--is-inside-work-tree"},
			Result: execx.Result{Stdout: []byte("true\n")},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"log", "--follow", "--reverse", "--format=%H%x09%s", "--", "internal/auth/session.go"},
			Result: execx.Result{
				Stdout: []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\tExtract session store (#98)\n"),
			},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s%n%n%b", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			Result: execx.Result{
				Stdout: []byte("Extract session store (#98)\n\nSplit session persistence away from auth handlers.\n"),
			},
		},
	)

	got, err := WhyFile(context.Background(), runner, "", "internal/auth/session.go")
	if err != nil {
		t.Fatalf("WhyFile() error = %v", err)
	}

	if len(got.Entries) != 1 {
		t.Fatalf("WhyFile() entries = %d, want 1", len(got.Entries))
	}
	if got.Entries[0].PullRequest == nil || got.Entries[0].PullRequest.Number != 98 {
		t.Fatalf("WhyFile() pull request = %#v, want PR #98", got.Entries[0].PullRequest)
	}
	if got.Entries[0].Summary.Source != "commit_body" {
		t.Fatalf("WhyFile() summary source = %q, want commit_body", got.Entries[0].Summary.Source)
	}
}
