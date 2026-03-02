package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/KalleBylin/chester/internal/execx"
)

type rangeEntry struct {
	PR     *PRDetails
	Direct bool
	SHA    string
	Why    string
}

func UnearthRange(ctx context.Context, runner execx.Runner, repo string, spec string) (string, error) {
	if err := RequireGitWorktree(ctx, runner); err != nil {
		return "", err
	}

	rows, err := GitRangeRows(ctx, runner, spec)
	if err != nil {
		return "", err
	}

	seenPRs := make(map[int]bool)
	prCache := make(map[int]PRDetails)
	entries := make([]rangeEntry, 0, len(rows))

	for _, row := range rows {
		number, hasPR, err := ResolveCommitPRNumber(ctx, runner, repo, row.SHA, row.Subject)
		if err != nil {
			return "", err
		}

		if hasPR {
			if seenPRs[number] {
				continue
			}
			seenPRs[number] = true

			details, ok := prCache[number]
			if !ok {
				details, err = LoadPRDetails(ctx, runner, repo, number)
				if err != nil {
					return "", err
				}
				prCache[number] = details
			}
			entries = append(entries, rangeEntry{
				PR:  &details,
				Why: PRWhy(details),
			})
			continue
		}

		message, err := GitCommitMessage(ctx, runner, row.SHA)
		if err != nil {
			return "", err
		}
		entries = append(entries, rangeEntry{
			Direct: true,
			SHA:    shortSHA(row.SHA),
			Why:    DirectCommitWhy(message),
		})
	}

	return renderUnearthRange(spec, entries), nil
}

func renderUnearthRange(spec string, entries []rangeEntry) string {
	var out strings.Builder

	fmt.Fprintf(&out, "# %s\n", spec)
	if len(entries) == 0 {
		out.WriteString("\n(none)\n")
		return out.String()
	}

	out.WriteString("\n")
	for i, entry := range entries {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("- ")
		if entry.Direct {
			fmt.Fprintf(&out, "direct %s: %s", entry.SHA, entry.Why)
			continue
		}
		fmt.Fprintf(&out, "PR #%d %s: %s", entry.PR.Number, entry.PR.Title, entry.Why)
	}

	return out.String()
}
