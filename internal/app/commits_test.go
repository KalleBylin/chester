package app

import (
	"context"
	"testing"

	"chester/internal/execx"
)

func TestResolveCommitPRNumberUsesCommitPullsEndpoint(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/commits/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/pulls",
			},
			Result: execx.Result{
				Stdout: []byte(`[{"number":98,"merged_at":"2026-02-15T18:00:00Z"}]`),
			},
		},
	)

	got, ok, err := ResolveCommitPRNumber(context.Background(), runner, "acme/chester", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "ignored")
	if err != nil {
		t.Fatalf("ResolveCommitPRNumber() error = %v", err)
	}
	if !ok || got != 98 {
		t.Fatalf("ResolveCommitPRNumber() = (%d, %v), want (98, true)", got, ok)
	}
}

func TestResolveCommitPRNumberFallsBackToMergeSubject(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/commits/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/pulls",
			},
			Result: execx.Result{
				Stdout: []byte(`[]`),
			},
		},
	)

	got, ok, err := ResolveCommitPRNumber(context.Background(), runner, "acme/chester", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "Merge pull request #98 from feature/session-store")
	if err != nil {
		t.Fatalf("ResolveCommitPRNumber() error = %v", err)
	}
	if !ok || got != 98 {
		t.Fatalf("ResolveCommitPRNumber() = (%d, %v), want (98, true)", got, ok)
	}
}

func TestResolveCommitPRNumberFallsBackToSquashSubject(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/commits/cccccccccccccccccccccccccccccccccccccccc/pulls",
			},
			Result: execx.Result{
				Stdout: []byte(`[]`),
			},
		},
	)

	got, ok, err := ResolveCommitPRNumber(context.Background(), runner, "acme/chester", "cccccccccccccccccccccccccccccccccccccccc", "Bypass ORM for hot path query (#77)")
	if err != nil {
		t.Fatalf("ResolveCommitPRNumber() error = %v", err)
	}
	if !ok || got != 77 {
		t.Fatalf("ResolveCommitPRNumber() = (%d, %v), want (77, true)", got, ok)
	}
}

func TestResolveCommitPRNumberReturnsDirectWhenNoPRExists(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "gh",
			Args: []string{
				"api", "-H", "Accept: application/vnd.github+json",
				"repos/acme/chester/commits/dddddddddddddddddddddddddddddddddddddddd/pulls",
			},
			Result: execx.Result{
				Stdout: []byte(`[]`),
			},
		},
	)

	got, ok, err := ResolveCommitPRNumber(context.Background(), runner, "acme/chester", "dddddddddddddddddddddddddddddddddddddddd", "Direct commit")
	if err != nil {
		t.Fatalf("ResolveCommitPRNumber() error = %v", err)
	}
	if ok || got != 0 {
		t.Fatalf("ResolveCommitPRNumber() = (%d, %v), want (0, false)", got, ok)
	}
}

func TestDirectCommitWhyPrefersBodyParagraph(t *testing.T) {
	t.Parallel()

	got := DirectCommitWhy("Fix nil panic\n\nFix nil panic when malformed cookie omits signature segment.")
	want := "Fix nil panic when malformed cookie omits signature segment."
	if got != want {
		t.Fatalf("DirectCommitWhy() = %q, want %q", got, want)
	}
}

func TestPRWhyFallsBackToTitleWhenBodyIsEmpty(t *testing.T) {
	t.Parallel()

	got := PRWhy(PRDetails{
		Title: "hotfix: ci will build tools & fix local_builder",
		Body:  "",
	})
	want := "hotfix: ci will build tools & fix local_builder"
	if got != want {
		t.Fatalf("PRWhy() = %q, want %q", got, want)
	}
}
