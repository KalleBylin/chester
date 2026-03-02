package cmd

import (
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
		Use:          "chester",
		Short:        "Deterministic repository archaeology for git and GitHub",
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVarP(&opts.Repo, "repo", "R", "", "override GitHub repo slug (owner/name)")

	root.AddCommand(newReadThreadCmd(opts))
	root.AddCommand(newFileHistoryCmd(opts))
	root.AddCommand(newUnearthLinesCmd(opts))
	root.AddCommand(newUnearthRangeCmd(opts))
	root.AddCommand(newOnboardCmd())

	return root
}
