package app

import "errors"

var (
	ErrGHNotInstalled    = errors.New("gh not found in PATH")
	ErrGHUnauthenticated = errors.New("gh is not authenticated for github.com")
	ErrGitNotInstalled   = errors.New("git not found in PATH")
	ErrNotGitRepository  = errors.New("not a git repository")
	ErrNoOriginRemote    = errors.New("repository has no origin remote")
	ErrNotGitHubRemote   = errors.New("origin remote is not a GitHub remote")
)
