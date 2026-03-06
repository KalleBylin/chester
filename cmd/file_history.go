package cmd

import (
	"github.com/KalleBylin/chester/internal/app"
	"github.com/spf13/cobra"
)

func newWhyFileCmd(opts *Options) *cobra.Command {
	var asJSON bool

	command := &cobra.Command{
		Use:          "why-file <path>",
		Aliases:      []string{"file-history"},
		Short:        "Show chronological history for one file, enriched with PR context when available",
		Example:      "chester why-file internal/auth/session.go",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := app.WhyFile(cmd.Context(), opts.Runner, app.MaybeResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo), args[0])
			if err != nil {
				return err
			}

			return writeCommandOutput(cmd, asJSON, app.RenderWhyFileMarkdown(result), result)
		},
	}

	command.Flags().BoolVar(&asJSON, "json", false, "render structured JSON instead of Markdown")
	return command
}
