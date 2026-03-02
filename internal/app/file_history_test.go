package app

import (
	"context"
	"strings"
	"testing"

	"chester/internal/execx"
	"chester/internal/testutil"
)

func TestFileHistoryRendersCollapsedTimeline(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "git",
			Args: []string{"rev-parse", "--is-inside-work-tree"},
			Result: execx.Result{Stdout: []byte("true\n")},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"log", "--follow", "--reverse", "--format=%H%x09%s", "--", "internal/auth/session.go"},
			Result: execx.Result{
				Stdout: []byte(strings.Join([]string{
					"1111111111111111111111111111111111111111\tIntroduce session store interface",
					"2222222222222222222222222222222222222222\tAdd invalidate helper",
					"3333333333333333333333333333333333333333\tFix nil panic",
				}, "\n")),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/1111111111111111111111111111111111111111/pulls"},
			Result: execx.Result{Stdout: []byte(`[{"number":98,"merged_at":"2026-02-15T18:00:00Z"}]`)},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"pr", "view", "98", "--repo", "acme/chester", "--json", "number,title,body,url,mergedAt"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pr_view_98.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/2222222222222222222222222222222222222222/pulls"},
			Result: execx.Result{Stdout: []byte(`[{"number":98,"merged_at":"2026-02-15T18:00:00Z"}]`)},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/3333333333333333333333333333333333333333/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s%n%n%b", "3333333333333333333333333333333333333333"},
			Result: execx.Result{
				Stdout: []byte("Fix nil panic\n\nFix nil panic when malformed cookie omits signature segment.\n"),
			},
		},
	)

	got, err := FileHistory(context.Background(), runner, "acme/chester", "internal/auth/session.go")
	if err != nil {
		t.Fatalf("FileHistory() error = %v", err)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "file_history.md")))
	if strings.TrimSpace(got) != want {
		t.Fatalf("FileHistory() mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestFileHistoryKeepsNonAdjacentEntriesSeparate(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "git",
			Args: []string{"rev-parse", "--is-inside-work-tree"},
			Result: execx.Result{Stdout: []byte("true\n")},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"log", "--follow", "--reverse", "--format=%H%x09%s", "--", "internal/auth/session.go"},
			Result: execx.Result{
				Stdout: []byte(strings.Join([]string{
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\tMerge pull request #98 from feature/session-store",
					"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\tDirect commit",
					"cccccccccccccccccccccccccccccccccccccccc\tAdd trailing cleanup (#98)",
				}, "\n")),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"pr", "view", "98", "--repo", "acme/chester", "--json", "number,title,body,url,mergedAt"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pr_view_98.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s%n%n%b", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
			Result: execx.Result{Stdout: []byte("Direct commit\n")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/cccccccccccccccccccccccccccccccccccccccc/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
	)

	got, err := FileHistory(context.Background(), runner, "acme/chester", "internal/auth/session.go")
	if err != nil {
		t.Fatalf("FileHistory() error = %v", err)
	}

	if strings.Count(got, "PR #98 Split session store") != 2 {
		t.Fatalf("FileHistory() collapsed non-adjacent entries:\n%s", got)
	}
}
