package cmd

import (
	"fmt"
	"strings"

	"github.com/KalleBylin/chester/internal/app"
	"github.com/spf13/cobra"
)

func newWhyRangeCmd(opts *Options) *cobra.Command {
	var asJSON bool

	command := &cobra.Command{
		Use:          "why-range <from_ref>..<to_ref>",
		Aliases:      []string{"unearth-range"},
		Short:        "Show chronological history for a git range, enriched with PR context when available",
		Example:      "chester why-range main..feature",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !strings.Contains(args[0], "..") {
				return fmt.Errorf("invalid range %q", args[0])
			}

			result, err := app.WhyRange(cmd.Context(), opts.Runner, app.MaybeResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo), args[0])
			if err != nil {
				return err
			}

			return writeCommandOutput(cmd, asJSON, app.RenderWhyRangeMarkdown(result), result)
		},
	}

	command.Flags().BoolVar(&asJSON, "json", false, "render structured JSON instead of Markdown")
	return command
}
