package cmd

import (
	"strings"

	"github.com/KalleBylin/chester/internal/execx"

	"github.com/spf13/cobra"
)

type Options struct {
	Repo   string
	Runner execx.Runner
}

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithOptions(&Options{
		Runner: execx.ExecRunner{},
	})
}

func NewRootCmdWithOptions(opts *Options) *cobra.Command {
	if opts == nil {
		opts = &Options{}
	}
	if opts.Runner == nil {
		opts.Runner = execx.ExecRunner{}
	}

	root := &cobra.Command{
		Use:   "chester",
		Short: "Deterministic repository archaeology for git and GitHub",
		Long:  "Chester helps coding agents answer why code exists by reading local git history and GitHub discussion without mutating the repository.",
		Example: strings.TrimSpace(`
chester file-history internal/auth/session.go
chester read-thread 123
chester unearth-lines db/queries.go:112:115
chester unearth-range main..feature`),
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVarP(&opts.Repo, "repo", "R", "", "override GitHub repo slug (owner/name)")

	root.AddCommand(newReadThreadCmd(opts))
	root.AddCommand(newFileHistoryCmd(opts))
	root.AddCommand(newUnearthLinesCmd(opts))
	root.AddCommand(newUnearthRangeCmd(opts))
	root.AddCommand(newOnboardCmd())
	root.InitDefaultCompletionCmd()

	return root
}
