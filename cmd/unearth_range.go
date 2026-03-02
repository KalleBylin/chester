package cmd

import (
	"fmt"
	"strings"

	"chester/internal/app"
	"github.com/spf13/cobra"
)

func newUnearthRangeCmd(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:          "unearth-range <from_ref>..<to_ref>",
		Short:        "Render a dense list of PRs represented by a git revision range",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !strings.Contains(args[0], "..") {
				return fmt.Errorf("invalid range %q", args[0])
			}

			repo, err := app.ResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo)
			if err != nil {
				return err
			}

			output, err := app.UnearthRange(cmd.Context(), opts.Runner, repo, args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), output)
			return err
		},
	}
}
