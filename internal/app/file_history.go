package app

import (
	"context"
	"fmt"
	"strings"

	"chester/internal/execx"
)

type fileHistoryEntry struct {
	SHAs   []string
	PR     *PRDetails
	Direct bool
	Why    string
}

func FileHistory(ctx context.Context, runner execx.Runner, repo string, path string) (string, error) {
	if err := RequireGitWorktree(ctx, runner); err != nil {
		return "", err
	}

	rows, err := GitFileHistoryRows(ctx, runner, path)
	if err != nil {
		return "", err
	}

	prCache := make(map[int]PRDetails)
	entries := make([]fileHistoryEntry, 0, len(rows))

	for _, row := range rows {
		number, hasPR, err := ResolveCommitPRNumber(ctx, runner, repo, row.SHA, row.Subject)
		if err != nil {
			return "", err
		}

		if hasPR {
			details, ok := prCache[number]
			if !ok {
				details, err = LoadPRDetails(ctx, runner, repo, number)
				if err != nil {
					return "", err
				}
				prCache[number] = details
			}

			short := shortSHA(row.SHA)
			if len(entries) > 0 && !entries[len(entries)-1].Direct && entries[len(entries)-1].PR != nil && entries[len(entries)-1].PR.Number == number {
				entries[len(entries)-1].SHAs = append(entries[len(entries)-1].SHAs, short)
				continue
			}

			entries = append(entries, fileHistoryEntry{
				SHAs: []string{short},
				PR:   &details,
				Why:  FirstParagraph(details.Body),
			})
			continue
		}

		message, err := GitCommitMessage(ctx, runner, row.SHA)
		if err != nil {
			return "", err
		}

		entries = append(entries, fileHistoryEntry{
			SHAs:   []string{shortSHA(row.SHA)},
			Direct: true,
			Why:    DirectCommitWhy(message),
		})
	}

	return renderFileHistory(path, entries), nil
}

func renderFileHistory(path string, entries []fileHistoryEntry) string {
	var out strings.Builder

	fmt.Fprintf(&out, "# %s\n", path)
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
		out.WriteString(strings.Join(entry.SHAs, ","))
		out.WriteString(" ")
		if entry.Direct {
			out.WriteString("direct\n")
		} else {
			fmt.Fprintf(&out, "PR #%d %s\n", entry.PR.Number, entry.PR.Title)
		}
		out.WriteString("  why: ")
		out.WriteString(entry.Why)
		out.WriteString("\n")
	}

	return out.String()
}
