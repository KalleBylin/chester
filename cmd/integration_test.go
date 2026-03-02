package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/KalleBylin/chester/internal/execx"
	"github.com/KalleBylin/chester/internal/testutil"
	"github.com/spf13/cobra"
)

func TestReadThreadCommandEndToEnd(t *testing.T) {
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

	stdout, stderr, err := executeForTest(t, runner, "-R", "acme/chester", "read-thread", "123")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "read_thread_pr.md")))
	if strings.TrimSpace(stdout) != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\n\ngot:\n%s", want, stdout)
	}
}

func TestFileHistoryCommandEndToEnd(t *testing.T) {
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

	stdout, stderr, err := executeForTest(t, runner, "-R", "acme/chester", "file-history", "internal/auth/session.go")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "file_history.md")))
	if strings.TrimSpace(stdout) != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\n\ngot:\n%s", want, stdout)
	}
}

func TestUnearthLinesCommandEndToEnd(t *testing.T) {
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

	stdout, stderr, err := executeForTest(t, runner, "-R", "acme/chester", "unearth-lines", "db/queries.go", "-L", "112,115")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "unearth_lines.md")))
	if strings.TrimSpace(stdout) != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\n\ngot:\n%s", want, stdout)
	}
}

func TestUnearthRangeCommandEndToEnd(t *testing.T) {
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

	stdout, stderr, err := executeForTest(t, runner, "-R", "acme/chester", "unearth-range", "main..feature")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	want := strings.TrimSpace(string(testutil.ReadFixture(t, "golden", "unearth_range.md")))
	if strings.TrimSpace(stdout) != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\n\ngot:\n%s", want, stdout)
	}
}

func TestOnboardCommandOutputsAgentSnippet(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := executeForTest(t, execx.NewMockRunner(), "onboard")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	wantFragments := []string{
		"chester Onboarding",
		"Chesterton's Fence",
		"`chester read-thread <id>`",
		"`chester --help`",
		".github/copilot-instructions.md",
	}
	for _, fragment := range wantFragments {
		if !strings.Contains(stdout, fragment) {
			t.Fatalf("stdout missing %q\n%s", fragment, stdout)
		}
	}
}

func TestCompletionCommandIsAvailable(t *testing.T) {
	t.Parallel()

	root := NewRootCmdWithOptions(&Options{Runner: execx.NewMockRunner()})
	completion := findChild(root, "completion")
	if completion == nil {
		t.Fatal("completion command missing")
	}

	for _, name := range []string{"bash", "zsh", "fish", "powershell"} {
		if findChild(completion, name) == nil {
			t.Fatalf("completion subcommand %q missing", name)
		}
	}
}

func executeForTest(t *testing.T, runner execx.Runner, args ...string) (string, string, error) {
	t.Helper()

	command := NewRootCmdWithOptions(&Options{
		Runner: runner,
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	bindOutputs(command, &stdout, &stderr)
	command.SetArgs(args)

	err := command.Execute()
	return stdout.String(), stderr.String(), err
}

func bindOutputs(command *cobra.Command, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	command.SetOut(stdout)
	command.SetErr(stderr)
	for _, child := range command.Commands() {
		bindOutputs(child, stdout, stderr)
	}
}

func findChild(command *cobra.Command, name string) *cobra.Command {
	for _, child := range command.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
