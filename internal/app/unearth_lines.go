package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/KalleBylin/chester/internal/execx"
)

type BlameSpan struct {
	SHA   string
	Start int
	End   int
}

type reviewCommentPayload struct {
	Path              string `json:"path"`
	Body              string `json:"body"`
	CreatedAt         string `json:"created_at"`
	Line              int    `json:"line"`
	OriginalLine      int    `json:"original_line"`
	StartLine         int    `json:"start_line"`
	OriginalStartLine int    `json:"original_start_line"`
	User              struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
}

func WhyLines(ctx context.Context, runner execx.Runner, repo string, file string, start int, end int) (WhyLinesResult, error) {
	if err := RequireGitWorktree(ctx, runner); err != nil {
		return WhyLinesResult{}, err
	}

	spans, err := GitBlameSpans(ctx, runner, file, start, end)
	if err != nil {
		return WhyLinesResult{}, err
	}

	prCache := make(map[int]PRDetails)
	noteCache := make(map[int][]ReviewComment)
	prRefs := make(map[int]PullRequestRef)
	noteOrder := make([]int, 0)
	results := make([]WhyLineSpanResult, 0, len(spans))

	for _, span := range spans {
		commit := NewCommitRef(span.SHA)
		prRef, hasPR, err := InferCommitPRRef(ctx, runner, repo, span.SHA, "")
		if err != nil {
			return WhyLinesResult{}, err
		}

		entry := WhyLineSpanResult{
			Start:  span.Start,
			End:    span.End,
			Commit: commit,
			Source: WhyLinesSource{Kind: "direct"},
		}

		if hasPR {
			pullRequest := PullRequestRef{
				Number: prRef.Number,
				Source: prRef.Source,
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
					pullRequest.Title = details.Title
					pullRequest.URL = details.URL
					pullRequest.Resolved = true
					entry.Summary = PRSummary(details)
				}

				if _, seen := noteCache[prRef.Number]; !seen {
					notes, ok := TryLoadPRReviewComments(ctx, runner, repo, prRef.Number, file)
					if ok {
						noteCache[prRef.Number] = notes
						if len(notes) > 0 {
							noteOrder = append(noteOrder, prRef.Number)
						}
					}
				}
			}
			if entry.Summary.Text == "" {
				message, err := GitCommitMessage(ctx, runner, span.SHA)
				if err != nil {
					return WhyLinesResult{}, err
				}
				entry.Summary = CommitSummary(message)
			}

			prRefs[prRef.Number] = pullRequest
			entry.Source = WhyLinesSource{
				Kind:        "pull_request",
				PullRequest: &pullRequest,
			}
			results = append(results, entry)
			continue
		}

		message, err := GitCommitMessage(ctx, runner, span.SHA)
		if err != nil {
			return WhyLinesResult{}, err
		}
		entry.Summary = CommitSummary(message)
		results = append(results, entry)
	}

	fileNotes := make([]WhyLinesNoteGroup, 0, len(noteOrder))
	for _, number := range noteOrder {
		comments := noteCache[number]
		if len(comments) == 0 {
			continue
		}
		fileNotes = append(fileNotes, WhyLinesNoteGroup{
			PullRequest: prRefs[number],
			Comments:    comments,
		})
	}

	return WhyLinesResult{
		File:      file,
		Start:     start,
		End:       end,
		Repo:      repo,
		Spans:     results,
		FileNotes: fileNotes,
	}, nil
}

func GitBlameSpans(ctx context.Context, runner execx.Runner, file string, start int, end int) ([]BlameSpan, error) {
	rangeArg := fmt.Sprintf("%d,%d", start, end)
	out, err := RunGit(ctx, runner, "blame", "--line-porcelain", "-L", rangeArg, "--", file)
	if err != nil {
		return nil, err
	}
	return parseBlameSpans(string(out))
}

func LoadPRReviewComments(ctx context.Context, runner execx.Runner, repo string, number int, file string) ([]ReviewComment, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/comments?per_page=100", repo, number)
	body, err := GHAPI(ctx, runner, "--paginate", "-H", "Accept: application/vnd.github+json", path)
	if err != nil {
		return nil, err
	}

	var payload []reviewCommentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	notes := make([]ReviewComment, 0, len(payload))
	for _, comment := range payload {
		if comment.Path != file {
			continue
		}
		if comment.User.Type == "Bot" {
			continue
		}
		cleanBody := SanitizeMarkdown(comment.Body)
		if cleanBody == "" {
			continue
		}
		startLine, endLine := commentRange(comment)
		notes = append(notes, ReviewComment{
			Author:    comment.User.Login,
			Body:      cleanBody,
			When:      comment.CreatedAt,
			Path:      comment.Path,
			StartLine: startLine,
			EndLine:   endLine,
		})
	}

	sort.SliceStable(notes, func(i, j int) bool {
		return notes[i].When < notes[j].When
	})
	if len(notes) > 3 {
		notes = notes[:3]
	}
	return notes, nil
}

func TryLoadPRReviewComments(ctx context.Context, runner execx.Runner, repo string, number int, file string) ([]ReviewComment, bool) {
	if repo == "" {
		return nil, false
	}
	comments, err := LoadPRReviewComments(ctx, runner, repo, number, file)
	if err != nil {
		return nil, false
	}
	return comments, true
}

func parseBlameSpans(output string) ([]BlameSpan, error) {
	type blamedLine struct {
		SHA  string
		Line int
	}

	rows := make([]blamedLine, 0)
	var currentSHA string
	var currentLine int
	haveHeader := false

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "\t") {
			if haveHeader {
				rows = append(rows, blamedLine{SHA: currentSHA, Line: currentLine})
				haveHeader = false
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 || !isLikelyHash(fields[0]) {
			continue
		}

		lineNo, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		currentSHA = fields[0]
		currentLine = lineNo
		haveHeader = true
	}

	if len(rows) == 0 {
		return nil, nil
	}

	spans := []BlameSpan{{
		SHA:   rows[0].SHA,
		Start: rows[0].Line,
		End:   rows[0].Line,
	}}

	for _, row := range rows[1:] {
		last := &spans[len(spans)-1]
		if last.SHA == row.SHA && row.Line == last.End+1 {
			last.End = row.Line
			continue
		}
		spans = append(spans, BlameSpan{
			SHA:   row.SHA,
			Start: row.Line,
			End:   row.Line,
		})
	}

	return spans, nil
}

func RenderWhyLinesMarkdown(result WhyLinesResult) string {
	var out strings.Builder

	fmt.Fprintf(&out, "# %s:%d-%d\n", result.File, result.Start, result.End)
	if len(result.Spans) == 0 {
		out.WriteString("\n(none)\n")
		return out.String()
	}

	out.WriteString("\n")
	for i, entry := range result.Spans {
		if i > 0 {
			out.WriteString("\n")
		}
		fmt.Fprintf(&out, "- %s\n", renderSpanLocation(entry.Start, entry.End))
		out.WriteString("  commit: ")
		out.WriteString(entry.Commit.Short)
		out.WriteString("\n")
		if entry.Source.Kind == "pull_request" && entry.Source.PullRequest != nil {
			out.WriteString("  source: ")
			out.WriteString(renderPullRequestHeading(*entry.Source.PullRequest))
			out.WriteString("\n")
		} else {
			out.WriteString("  source: direct\n")
		}
		out.WriteString("  summary: ")
		out.WriteString(entry.Summary.Text)
		out.WriteString("\n")
		out.WriteString("  summary-source: ")
		out.WriteString(renderSummarySource(entry.Summary.Source))
		out.WriteString("\n")
	}

	if len(result.FileNotes) > 0 {
		out.WriteString("\n## File Notes\n")
		for i, group := range result.FileNotes {
			if i > 0 {
				out.WriteString("\n")
			}
			out.WriteString("- ")
			out.WriteString(renderPullRequestHeading(group.PullRequest))
			out.WriteString("\n")
			for _, comment := range group.Comments {
				out.WriteString("  ")
				out.WriteString(renderReviewCommentLabel(comment))
				out.WriteString(": ")
				out.WriteString(comment.Body)
				out.WriteString("\n")
			}
		}
	}

	return strings.TrimRight(out.String(), "\n")
}

func renderSpanLocation(start int, end int) string {
	if start == end {
		return fmt.Sprintf("L%d", start)
	}
	return fmt.Sprintf("L%d-L%d", start, end)
}

func isLikelyHash(value string) bool {
	if len(value) < 8 {
		return false
	}
	for i := 0; i < len(value); i++ {
		switch {
		case value[i] >= '0' && value[i] <= '9':
		case value[i] >= 'a' && value[i] <= 'f':
		default:
			return false
		}
	}
	return true
}

func commentRange(comment reviewCommentPayload) (int, int) {
	start := comment.StartLine
	if start == 0 {
		start = comment.OriginalStartLine
	}
	end := comment.Line
	if end == 0 {
		end = comment.OriginalLine
	}
	if start == 0 {
		start = end
	}
	if end == 0 {
		end = start
	}
	return start, end
}

func renderReviewCommentLabel(comment ReviewComment) string {
	label := "@" + comment.Author
	switch {
	case comment.StartLine > 0 && comment.EndLine > 0 && comment.StartLine != comment.EndLine:
		return fmt.Sprintf("%s (L%d-L%d)", label, comment.StartLine, comment.EndLine)
	case comment.StartLine > 0:
		return fmt.Sprintf("%s (L%d)", label, comment.StartLine)
	default:
		return label
	}
}
