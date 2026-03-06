package cmd

import (
	"github.com/KalleBylin/chester/internal/app"
	"github.com/spf13/cobra"
)

func newReadThreadCmd(opts *Options) *cobra.Command {
	var asJSON bool

	command := &cobra.Command{
		Use:          "read-thread <id>",
		Short:        "Fetch an issue or PR thread and render structured human discussion",
		Example:      "chester read-thread 123",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := app.ResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo)
			if err != nil {
				return err
			}

			result, err := app.ReadThread(cmd.Context(), opts.Runner, repo, args[0])
			if err != nil {
				return err
			}

			return writeCommandOutput(cmd, asJSON, app.RenderReadThreadMarkdown(result), result)
		},
	}

	command.Flags().BoolVar(&asJSON, "json", false, "render structured JSON instead of Markdown")
	return command
}
