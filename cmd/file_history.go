package cmd

import (
	"fmt"

	"github.com/KalleBylin/chester/internal/app"
	"github.com/spf13/cobra"
)

func newFileHistoryCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:          "file-history <path>",
		Short:        "Show chronological PR-backed history for one file",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := app.ResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo)
			if err != nil {
				return err
			}

			output, err := app.FileHistory(cmd.Context(), opts.Runner, repo, args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), output)
			return err
		},
	}
}
