package app

import (
	"context"
	"errors"
	"testing"

	"github.com/KalleBylin/chester/internal/execx"
)

func TestResolveRepoSlugUsesOverride(t *testing.T) {
	t.Parallel()

	got, err := ResolveRepoSlug(context.Background(), execx.NewMockRunner(), "acme/chester")
	if err != nil {
		t.Fatalf("ResolveRepoSlug() error = %v", err)
	}
	if got != "acme/chester" {
		t.Fatalf("ResolveRepoSlug() = %q", got)
	}
}

func TestResolveRepoSlugReadsOriginRemote(t *testing.T) {
	t.Parallel()

	runner := execx.NewMockRunner(
		execx.Expectation{
			Name: "git",
			Args: []string{"remote", "get-url", "origin"},
			Result: execx.Result{
				Stdout: []byte("git@github.com:acme/chester.git\n"),
			},
		},
	)

	got, err := ResolveRepoSlug(context.Background(), runner, "")
	if err != nil {
		t.Fatalf("ResolveRepoSlug() error = %v", err)
	}
	if got != "acme/chester" {
		t.Fatalf("ResolveRepoSlug() = %q", got)
	}
}

func TestResolveRepoSlugMapsCommonGitFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "not git repo",
			err: &execx.RunError{
				Name:   "git",
				Args:   []string{"remote", "get-url", "origin"},
				Stderr: []byte("fatal: not a git repository (or any of the parent directories): .git"),
			},
			want: ErrNotGitRepository,
		},
		{
			name: "missing origin",
			err: &execx.RunError{
				Name:   "git",
				Args:   []string{"remote", "get-url", "origin"},
				Stderr: []byte("error: No such remote 'origin'"),
			},
			want: ErrNoOriginRemote,
		},
		{
			name: "missing git binary",
			err:  ErrCommandNotFoundWrapped(),
			want: ErrGitNotInstalled,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := execx.NewMockRunner(
				execx.Expectation{
					Name: "git",
					Args: []string{"remote", "get-url", "origin"},
					Err:  tt.err,
				},
			)

			_, err := ResolveRepoSlug(context.Background(), runner, "")
			if !errors.Is(err, tt.want) {
				t.Fatalf("ResolveRepoSlug() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseGitHubRepoSlugSupportsKnownRemoteFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		remote string
		want   string
	}{
		{remote: "git@github.com:acme/chester.git", want: "acme/chester"},
		{remote: "https://github.com/acme/chester.git", want: "acme/chester"},
		{remote: "ssh://git@github.com/acme/chester.git", want: "acme/chester"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.remote, func(t *testing.T) {
			t.Parallel()

			got, err := ParseGitHubRepoSlug(tt.remote)
			if err != nil {
				t.Fatalf("ParseGitHubRepoSlug() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseGitHubRepoSlug() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseGitHubRepoSlugRejectsUnsupportedRemotes(t *testing.T) {
	t.Parallel()

	_, err := ParseGitHubRepoSlug("git@gitlab.com:acme/chester.git")
	if !errors.Is(err, ErrNotGitHubRemote) {
		t.Fatalf("ParseGitHubRepoSlug() error = %v, want ErrNotGitHubRemote", err)
	}
}

func ErrCommandNotFoundWrapped() error {
	return errors.Join(execx.ErrCommandNotFound, errors.New("git"))
}
