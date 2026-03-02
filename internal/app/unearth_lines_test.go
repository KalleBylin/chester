package app

import (
	"context"
	"strings"
	"testing"

	"chester/internal/execx"
	"chester/internal/testutil"
)

func TestParseBlameSpansCollapsesContiguousLines(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa 1 112 2",
		"author Alice",
		"\tfirst line",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa 2 113",
		"author Alice",
		"\tsecond line",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb 3 114 2",
		"author Bob",
		"\tthird line",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb 4 115",
		"author Bob",
		"\tfourth line",
	}, "\n")

	got, err := parseBlameSpans(input)
	if err != nil {
		t.Fatalf("parseBlameSpans() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("parseBlameSpans() len = %d, want 2", len(got))
	}
	if got[0].Start != 112 || got[0].End != 113 || got[1].Start != 114 || got[1].End != 115 {
		t.Fatalf("parseBlameSpans() = %#v", got)
	}
}

func TestLoadPRReviewNotesFiltersByFileAndLimitsToThree(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "--paginate", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/pulls/77/comments?per_page=100",
			},
			Result: execx.Result{
				Stdout: testutil.ReadFixture(t, "gh", "pull_comments_77_many.json"),
			},
		},
	)

	got, err := LoadPRReviewNotes(context.Background(), runner, "acme/chester", 77, "db/queries.go")
	if err != nil {
		t.Fatalf("LoadPRReviewNotes() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("LoadPRReviewNotes() len = %d, want 3", len(got))
	}
	if got[0].Author != "maintainer" || got[1].Author != "reviewer" || got[2].Author != "architect" {
		t.Fatalf("LoadPRReviewNotes() = %#v", got)
	}
}

func TestUnearthLinesRendersPRAndDirectFallback(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "git",
			Args: []string{"rev-parse", "--is-inside-work-tree"},
			Result: execx.Result{Stdout: []byte("true\n")},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"blame", "--line-porcelain", "-L", "112,115", "--", "db/queries.go"},
			Result: execx.Result{
				Stdout: []byte(strings.Join([]string{
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa 1 112 2",
					"author Alice",
					"\tfirst line",
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa 2 113",
					"author Alice",
					"\tsecond line",
					"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb 3 114 2",
					"author Bob",
					"\tthird line",
					"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb 4 115",
					"author Bob",
					"\tfourth line",
				}, "\n")),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			Result: execx.Result{Stdout: []byte("Bypass ORM for hot path query (#77)\n")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"pr", "view", "77", "--repo", "acme/chester", "--json", "number,title,body,url,mergedAt"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pr_view_77.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "--paginate", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/pulls/77/comments?per_page=100"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pull_comments_77.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
			Result: execx.Result{Stdout: []byte("Direct fix\n")},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s%n%n%b", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
			Result: execx.Result{
				Stdout: []byte("Direct fix\n\nNormalize placeholders so the query stays portable across drivers.\n"),
			},
		},
	)

	got, err := UnearthLines(context.Background(), runner, "acme/chester", "db/queries.go", 112, 115)
	if err != nil {
		t.Fatalf("UnearthLines() error = %v", err)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "unearth_lines.md")))
	if strings.TrimSpace(got) != want {
		t.Fatalf("UnearthLines() mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}
