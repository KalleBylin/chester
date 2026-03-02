package cmd

import (
	"fmt"

	"chester/internal/app"
	"github.com/spf13/cobra"
)

func newReadThreadCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:          "read-thread <id>",
		Short:        "Fetch an issue or PR thread and render a compact conversation transcript",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := app.ResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo)
			if err != nil {
				return err
			}

			output, err := app.ReadThread(cmd.Context(), opts.Runner, repo, args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), output)
			return err
		},
	}
}
