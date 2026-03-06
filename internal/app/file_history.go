package app

import (
	"context"

	"github.com/KalleBylin/chester/internal/execx"
)

func WhyFile(ctx context.Context, runner execx.Runner, repo string, path string) (WhyFileResult, error) {
	if err := RequireGitWorktree(ctx, runner); err != nil {
		return WhyFileResult{}, err
	}

	rows, err := GitFileHistoryRows(ctx, runner, path)
	if err != nil {
		return WhyFileResult{}, err
	}

	prCache := make(map[int]PRDetails)
	entries := make([]HistoryEntry, 0, len(rows))

	for _, row := range rows {
		commit := NewCommitRef(row.SHA)
		prRef, hasPR, err := InferCommitPRRef(ctx, runner, repo, row.SHA, row.Subject)
		if err != nil {
			return WhyFileResult{}, err
		}

		if hasPR {
			if len(entries) > 0 && entries[len(entries)-1].PullRequest != nil && entries[len(entries)-1].PullRequest.Number == prRef.Number {
				entries[len(entries)-1].Commits = append(entries[len(entries)-1].Commits, commit)
				continue
			}

			entry := HistoryEntry{
				Commits: []CommitRef{commit},
				PullRequest: &PullRequestRef{
					Number: prRef.Number,
					Source: prRef.Source,
				},
			}

			if repo != "" {
				details, ok := prCache[prRef.Number]
				if !ok {
					var loaded bool
					details, loaded = TryLoadPRDetails(ctx, runner, repo, prRef.Number)
					if loaded {
						prCache[prRef.Number] = details
						ok = true
					}
				}
				if ok {
					entry.PullRequest.Title = details.Title
					entry.PullRequest.URL = details.URL
					entry.PullRequest.Resolved = true
					entry.Summary = PRSummary(details)
				}
			}
			if entry.Summary.Text == "" {
				message, err := GitCommitMessage(ctx, runner, row.SHA)
				if err != nil {
					return WhyFileResult{}, err
				}
				entry.Summary = CommitSummary(message)
			}

			entries = append(entries, entry)
			continue
		}

		message, err := GitCommitMessage(ctx, runner, row.SHA)
		if err != nil {
			return WhyFileResult{}, err
		}

		entries = append(entries, HistoryEntry{
			Commits: []CommitRef{commit},
			Summary: CommitSummary(message),
		})
	}

	return WhyFileResult{
		Path:    path,
		Repo:    repo,
		Entries: entries,
	}, nil
}

func RenderWhyFileMarkdown(result WhyFileResult) string {
	return renderHistoryMarkdown(result.Path, "commits", result.Entries)
}
