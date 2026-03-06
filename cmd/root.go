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
		Long:  "Chester helps coding agents answer why code exists by reading local git history first, then layering in GitHub discussion when it is available.",
		Example: strings.TrimSpace(`
chester why-file internal/auth/session.go
chester read-thread 123
chester why-lines db/queries.go:112:115
chester why-range main..feature
chester text-history "SessionStore"`),
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVarP(&opts.Repo, "repo", "R", "", "override GitHub repo slug (owner/name)")

	root.AddCommand(newReadThreadCmd(opts))
	root.AddCommand(newWhyFileCmd(opts))
	root.AddCommand(newWhyLinesCmd(opts))
	root.AddCommand(newWhyRangeCmd(opts))
	root.AddCommand(newTextHistoryCmd(opts))
	root.AddCommand(newOnboardCmd())
	root.InitDefaultCompletionCmd()

	return root
}
