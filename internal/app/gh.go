package app

import (
	"context"
	"errors"
	"strings"

	"chester/internal/execx"
)

func RunGH(ctx context.Context, runner execx.Runner, args ...string) ([]byte, error) {
	result, err := runner.Run(ctx, "gh", args...)
	if err != nil {
		if errors.Is(err, execx.ErrCommandNotFound) {
			return nil, ErrGHNotInstalled
		}

		var runErr *execx.RunError
		if errors.As(err, &runErr) && isGHAuthFailure(string(runErr.Stderr)) {
			return nil, ErrGHUnauthenticated
		}
		return nil, err
	}
	return result.Stdout, nil
}

func GHPRView(ctx context.Context, runner execx.Runner, repo string, id string, fields []string) ([]byte, error) {
	args := []string{"pr", "view", id, "--repo", repo, "--json", strings.Join(fields, ",")}
	return RunGH(ctx, runner, args...)
}

func GHIssueView(ctx context.Context, runner execx.Runner, repo string, id string, fields []string) ([]byte, error) {
	args := []string{"issue", "view", id, "--repo", repo, "--json", strings.Join(fields, ",")}
	return RunGH(ctx, runner, args...)
}

func GHAPI(ctx context.Context, runner execx.Runner, args ...string) ([]byte, error) {
	callArgs := append([]string{"api"}, args...)
	return RunGH(ctx, runner, callArgs...)
}

func isGHAuthFailure(stderr string) bool {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return false
	}

	known := []string{
		"gh auth login",
		"not logged into any GitHub hosts",
		"authentication failed",
	}
	for _, fragment := range known {
		if strings.Contains(stderr, fragment) {
			return true
		}
	}
	return false
}
