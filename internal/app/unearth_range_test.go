package app

import (
	"context"
	"strings"
	"testing"

	"github.com/KalleBylin/chester/internal/execx"
	"github.com/KalleBylin/chester/internal/testutil"
)

func TestUnearthRangeDedupesPRsByFirstSeenOrder(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "git",
			Args: []string{"rev-parse", "--is-inside-work-tree"},
			Result: execx.Result{Stdout: []byte("true\n")},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"log", "--reverse", "--format=%H%x09%s", "main..feature"},
			Result: execx.Result{
				Stdout: []byte(strings.Join([]string{
					"1111111111111111111111111111111111111111\tExtract session store (#98)",
					"2222222222222222222222222222222222222222\tAnother commit for the same PR (#98)",
					"3333333333333333333333333333333333333333\tInvalidate sessions on password reset (#151)",
					"4444444444444444444444444444444444444444\tRegenerate golden CLI output",
				}, "\n")),
			},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/1111111111111111111111111111111111111111/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"pr", "view", "98", "--repo", "acme/chester", "--json", "number,title,body,url,mergedAt"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pr_view_98.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/2222222222222222222222222222222222222222/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/3333333333333333333333333333333333333333/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"pr", "view", "151", "--repo", "acme/chester", "--json", "number,title,body,url,mergedAt"},
			Result: execx.Result{Stdout: testutil.ReadFixture(t, "gh", "pr_view_151.json")},
		},
		execx.Expectation{
			Name: "gh",
			Args: []string{"api", "-H", "Accept: application/vnd.github+json", "repos/acme/chester/commits/4444444444444444444444444444444444444444/pulls"},
			Result: execx.Result{Stdout: []byte(`[]`)},
		},
		execx.Expectation{
			Name: "git",
			Args: []string{"show", "-s", "--format=%s%n%n%b", "4444444444444444444444444444444444444444"},
			Result: execx.Result{
				Stdout: []byte("Regenerate golden CLI output\n\nRegenerate golden CLI output for the auth fixtures.\n"),
			},
		},
	)

	got, err := UnearthRange(context.Background(), runner, "acme/chester", "main..feature")
	if err != nil {
		t.Fatalf("UnearthRange() error = %v", err)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "unearth_range.md")))
	if strings.TrimSpace(got) != want {
		t.Fatalf("UnearthRange() mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}
