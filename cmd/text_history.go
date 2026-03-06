package cmd

import (
	"github.com/KalleBylin/chester/internal/app"
	"github.com/spf13/cobra"
)

func newTextHistoryCmd(opts *Options) *cobra.Command {
	var asJSON bool
	var path string

	command := &cobra.Command{
		Use:          "text-history <literal>",
		Short:        "Show chronological history for one exact text literal",
		Example:      `chester text-history "SessionStore"`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.TextHistory(cmd.Context(), opts.Runner, app.MaybeResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo), args[0], path)
			if err != nil {
				return err
			}

			return writeCommandOutput(cmd, asJSON, app.RenderTextHistoryMarkdown(result), result)
		},
	}

	command.Flags().BoolVar(&asJSON, "json", false, "render structured JSON instead of Markdown")
	command.Flags().StringVar(&path, "path", "", "restrict the search to one exact path")
	return command
}
