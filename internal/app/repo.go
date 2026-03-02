package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"chester/internal/execx"
)

func ResolveRepoSlug(ctx context.Context, runner execx.Runner, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	result, err := runner.Run(ctx, "git", "remote", "get-url", "origin")
	if err != nil {
		if errors.Is(err, execx.ErrCommandNotFound) {
			return "", ErrGitNotInstalled
		}

		var runErr *execx.RunError
		if errors.As(err, &runErr) {
			stderr := strings.TrimSpace(string(runErr.Stderr))
			switch {
			case strings.Contains(stderr, "not a git repository"):
				return "", ErrNotGitRepository
			case strings.Contains(stderr, "No such remote 'origin'"):
				return "", ErrNoOriginRemote
			}
		}
		return "", err
	}

	slug, parseErr := ParseGitHubRepoSlug(strings.TrimSpace(string(result.Stdout)))
	if parseErr != nil {
		return "", parseErr
	}
	return slug, nil
}

func ParseGitHubRepoSlug(remoteURL string) (string, error) {
	remoteURL = strings.TrimSpace(remoteURL)
	switch {
	case strings.HasPrefix(remoteURL, "git@github.com:"):
		return trimGitHubSlug(strings.TrimPrefix(remoteURL, "git@github.com:"))
	case strings.HasPrefix(remoteURL, "https://github.com/"):
		return trimGitHubSlug(strings.TrimPrefix(remoteURL, "https://github.com/"))
	case strings.HasPrefix(remoteURL, "ssh://git@github.com/"):
		return trimGitHubSlug(strings.TrimPrefix(remoteURL, "ssh://git@github.com/"))
	default:
		return "", ErrNotGitHubRemote
	}
}

func trimGitHubSlug(value string) (string, error) {
	value = strings.TrimSuffix(value, ".git")
	value = strings.Trim(value, "/")
	parts := strings.Split(value, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("%w: %s", ErrNotGitHubRemote, value)
	}
	return parts[0] + "/" + parts[1], nil
}
