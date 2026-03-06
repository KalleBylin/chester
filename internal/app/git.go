package app

import (
	"context"
	"errors"
	"strings"

	"github.com/KalleBylin/chester/internal/execx"
)

type CommitRow struct {
	SHA     string
	Subject string
}

func RunGit(ctx context.Context, runner execx.Runner, args ...string) ([]byte, error) {
	result, err := runner.Run(ctx, "git", args...)
	if err != nil {
		if errors.Is(err, execx.ErrCommandNotFound) {
			return nil, ErrGitNotInstalled
		}

		var runErr *execx.RunError
		if errors.As(err, &runErr) {
			stderr := strings.TrimSpace(string(runErr.Stderr))
			if strings.Contains(stderr, "not a git repository") {
				return nil, ErrNotGitRepository
			}
		}
		return nil, err
	}
	return result.Stdout, nil
}

func RequireGitWorktree(ctx context.Context, runner execx.Runner) error {
	out, err := RunGit(ctx, runner, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) != "true" {
		return ErrNotGitRepository
	}
	return nil
}

func GitFileHistoryRows(ctx context.Context, runner execx.Runner, path string) ([]CommitRow, error) {
	out, err := RunGit(ctx, runner, "log", "--follow", "--reverse", "--format=%H%x09%s", "--", path)
	if err != nil {
		return nil, err
	}
	return parseCommitRows(string(out)), nil
}

func GitRangeRows(ctx context.Context, runner execx.Runner, spec string) ([]CommitRow, error) {
	out, err := RunGit(ctx, runner, "log", "--reverse", "--format=%H%x09%s", spec)
	if err != nil {
		return nil, err
	}
	return parseCommitRows(string(out)), nil
}

func GitTextHistoryRows(ctx context.Context, runner execx.Runner, literal string, path string) ([]CommitRow, error) {
	args := []string{"log", "--reverse", "--format=%H%x09%s", "-S", literal}
	if path != "" {
		args = append(args, "--", path)
	}
	out, err := RunGit(ctx, runner, args...)
	if err != nil {
		return nil, err
	}
	return parseCommitRows(string(out)), nil
}

func GitCommitSubject(ctx context.Context, runner execx.Runner, sha string) (string, error) {
	out, err := RunGit(ctx, runner, "show", "-s", "--format=%s", sha)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func GitCommitMessage(ctx context.Context, runner execx.Runner, sha string) (string, error) {
	out, err := RunGit(ctx, runner, "show", "-s", "--format=%s%n%n%b", sha)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func parseCommitRows(output string) []CommitRow {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	rows := make([]CommitRow, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		row := CommitRow{SHA: parts[0]}
		if len(parts) == 2 {
			row.Subject = parts[1]
		}
		rows = append(rows, row)
	}
	return rows
}
