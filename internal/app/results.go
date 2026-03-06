package app

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Summary struct {
	Text   string `json:"text"`
	Source string `json:"source"`
}

type CommitRef struct {
	SHA   string `json:"sha"`
	Short string `json:"short"`
}

type PullRequestRef struct {
	Number   int    `json:"number"`
	Title    string `json:"title,omitempty"`
	URL      string `json:"url,omitempty"`
	Resolved bool   `json:"resolved"`
	Source   string `json:"source"`
}

type HistoryEntry struct {
	Commits     []CommitRef     `json:"commits"`
	PullRequest *PullRequestRef `json:"pull_request,omitempty"`
	Summary     Summary         `json:"summary"`
}

type WhyFileResult struct {
	Path    string         `json:"path"`
	Repo    string         `json:"repo,omitempty"`
	Entries []HistoryEntry `json:"entries"`
}

type WhyRangeResult struct {
	Spec    string         `json:"spec"`
	Repo    string         `json:"repo,omitempty"`
	Entries []HistoryEntry `json:"entries"`
}

type TextHistoryResult struct {
	Literal string         `json:"literal"`
	Path    string         `json:"path,omitempty"`
	Repo    string         `json:"repo,omitempty"`
	Entries []HistoryEntry `json:"entries"`
}

type WhyLinesSource struct {
	Kind        string          `json:"kind"`
	PullRequest *PullRequestRef `json:"pull_request,omitempty"`
}

type WhyLineSpanResult struct {
	Start   int            `json:"start"`
	End     int            `json:"end"`
	Commit  CommitRef      `json:"commit"`
	Source  WhyLinesSource `json:"source"`
	Summary Summary        `json:"summary"`
}

type ReviewComment struct {
	Author    string `json:"author"`
	Body      string `json:"body"`
	When      string `json:"when"`
	Path      string `json:"path,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

type WhyLinesNoteGroup struct {
	PullRequest PullRequestRef  `json:"pull_request"`
	Comments    []ReviewComment `json:"comments"`
}

type WhyLinesResult struct {
	File      string              `json:"file"`
	Start     int                 `json:"start"`
	End       int                 `json:"end"`
	Repo      string              `json:"repo,omitempty"`
	Spans     []WhyLineSpanResult `json:"spans"`
	FileNotes []WhyLinesNoteGroup `json:"file_notes,omitempty"`
}

type ThreadComment struct {
	Author string `json:"author"`
	Body   string `json:"body"`
	When   string `json:"when"`
}

type ThreadReview struct {
	Author string `json:"author"`
	State  string `json:"state"`
	Body   string `json:"body"`
	When   string `json:"when"`
}

type ReadThreadResult struct {
	Number         int             `json:"number"`
	Kind           string          `json:"kind"`
	Title          string          `json:"title"`
	URL            string          `json:"url"`
	Body           string          `json:"body"`
	Comments       []ThreadComment `json:"comments"`
	Reviews        []ThreadReview  `json:"reviews,omitempty"`
	ReviewComments []ReviewComment `json:"review_comments,omitempty"`
}

func RenderJSON(value any) (string, error) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func renderHistoryMarkdown(heading string, label string, entries []HistoryEntry) string {
	var out strings.Builder

	fmt.Fprintf(&out, "# %s\n", heading)
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
		if entry.PullRequest == nil {
			out.WriteString("direct\n")
			out.WriteString("  commit: ")
			out.WriteString(entry.Commits[0].Short)
			out.WriteString("\n")
		} else {
			out.WriteString(renderPullRequestHeading(*entry.PullRequest))
			out.WriteString("\n")
			out.WriteString("  ")
			out.WriteString(label)
			out.WriteString(": ")
			out.WriteString(renderCommits(entry.Commits))
			out.WriteString("\n")
		}
		out.WriteString("  summary: ")
		out.WriteString(entry.Summary.Text)
		out.WriteString("\n")
		out.WriteString("  summary-source: ")
		out.WriteString(renderSummarySource(entry.Summary.Source))
		out.WriteString("\n")
	}

	return out.String()
}

func renderPullRequestHeading(pr PullRequestRef) string {
	if pr.Title == "" {
		return fmt.Sprintf("PR #%d", pr.Number)
	}
	return fmt.Sprintf("PR #%d %s", pr.Number, pr.Title)
}

func renderCommits(commits []CommitRef) string {
	parts := make([]string, 0, len(commits))
	for _, commit := range commits {
		parts = append(parts, commit.Short)
	}
	return strings.Join(parts, ",")
}

func renderSummarySource(source string) string {
	switch source {
	case "pr_body":
		return "pr body"
	case "pr_title":
		return "pr title"
	case "commit_body":
		return "commit body"
	case "commit_subject":
		return "commit subject"
	default:
		return source
	}
}
