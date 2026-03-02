package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"chester/internal/app"
	"github.com/spf13/cobra"
)

func newUnearthLinesCmd(opts *Options) *cobra.Command {
	var lineSpec string

	command := &cobra.Command{
		Use:          "unearth-lines <file>",
		Short:        "Explain why an exact blamed line range exists",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if lineSpec == "" {
				return fmt.Errorf(`required flag(s) "lines" not set`)
			}

			start, end, err := parseLineSpec(lineSpec)
			if err != nil {
				return err
			}

			repo, err := app.ResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo)
			if err != nil {
				return err
			}

			output, err := app.UnearthLines(cmd.Context(), opts.Runner, repo, args[0], start, end)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), output)
			return err
		},
	}

	command.Flags().StringVarP(&lineSpec, "lines", "L", "", "line range in start,end form")
	return command
}

func parseLineSpec(value string) (int, int, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid line range %q", value)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line range %q", value)
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line range %q", value)
	}
	if start <= 0 || end < start {
		return 0, 0, fmt.Errorf("invalid line range %q", value)
	}
	return start, end, nil
}
