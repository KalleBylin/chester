package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/KalleBylin/chester/internal/app"
	"github.com/spf13/cobra"
)

var (
	inlineColonLineSpec = regexp.MustCompile(`^(.+):(\d+):(\d+)$`)
	inlineDashLineSpec  = regexp.MustCompile(`^(.+):(\d+)-(\d+)$`)
)

func newUnearthLinesCmd(opts *Options) *cobra.Command {
	var lineSpec string

	command := &cobra.Command{
		Use:          "unearth-lines <file>:<start>:<end>",
		Short:        "Explain why an exact blamed line range exists",
		Long:         "Use an editor-style location to point at the exact lines you want to explain. The legacy --lines flag is still accepted for compatibility, but the inline form is the primary syntax.",
		Example:      "chester unearth-lines db/queries.go:112:115",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, start, end, err := parseLineTarget(args[0], lineSpec)
			if err != nil {
				return err
			}

			repo, err := app.ResolveRepoSlug(cmd.Context(), opts.Runner, opts.Repo)
			if err != nil {
				return err
			}

			output, err := app.UnearthLines(cmd.Context(), opts.Runner, repo, file, start, end)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), output)
			return err
		},
	}

	command.Flags().StringVarP(&lineSpec, "lines", "L", "", "line range in start,end form")
	_ = command.Flags().MarkHidden("lines")
	return command
}

func parseLineTarget(arg string, lineSpec string) (string, int, int, error) {
	if lineSpec != "" {
		if _, _, _, ok, err := parseInlineLineSpec(arg); err != nil {
			return "", 0, 0, err
		} else if ok {
			return "", 0, 0, fmt.Errorf("use either an inline location or --lines, not both")
		}

		start, end, err := parseLineSpec(lineSpec)
		if err != nil {
			return "", 0, 0, err
		}
		return arg, start, end, nil
	}

	file, start, end, ok, err := parseInlineLineSpec(arg)
	if err != nil {
		return "", 0, 0, err
	}
	if ok {
		return file, start, end, nil
	}

	return "", 0, 0, fmt.Errorf("invalid line target %q (want <file>:<start>:<end>)", arg)
}

func parseInlineLineSpec(value string) (string, int, int, bool, error) {
	for _, pattern := range []*regexp.Regexp{inlineColonLineSpec, inlineDashLineSpec} {
		matches := pattern.FindStringSubmatch(value)
		if matches == nil {
			continue
		}

		start, end, err := parseLineBounds(matches[2], matches[3], value)
		if err != nil {
			return "", 0, 0, false, err
		}
		return matches[1], start, end, true, nil
	}
	return "", 0, 0, false, nil
}

func parseLineSpec(value string) (int, int, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid line range %q", value)
	}

	return parseLineBounds(parts[0], parts[1], value)
}

func parseLineBounds(startValue string, endValue string, original string) (int, int, error) {
	start, err := strconv.Atoi(strings.TrimSpace(startValue))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line range %q", original)
	}
	end, err := strconv.Atoi(strings.TrimSpace(endValue))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line range %q", original)
	}
	if start <= 0 || end < start {
		return 0, 0, fmt.Errorf("invalid line range %q", original)
	}
	return start, end, nil
}
