package app

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"chester/internal/execx"
)

type BlameSpan struct {
	SHA   string
	Start int
	End   int
}

type ReviewNote struct {
	Author string
	Body   string
	When   string
}

type reviewCommentPayload struct {
	Path      string `json:"path"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
}

func UnearthLines(ctx context.Context, runner execx.Runner, repo string, file string, start int, end int) (string, error) {
	if err := RequireGitWorktree(ctx, runner); err != nil {
		return "", err
	}

	spans, err := GitBlameSpans(ctx, runner, file, start, end)
	if err != nil {
		return "", err
	}

	prCache := make(map[int]PRDetails)
	noteCache := make(map[int][]ReviewNote)
	lines := make([]string, 0, len(spans))

	for _, span := range spans {
		number, hasPR, err := ResolveCommitPRNumber(ctx, runner, repo, span.SHA, "")
		if err != nil {
			return "", err
		}

		header := "- " + renderBlameSpan(span) + " " + shortSHA(span.SHA) + " "
		if hasPR {
			details, ok := prCache[number]
			if !ok {
				details, err = LoadPRDetails(ctx, runner, repo, number)
				if err != nil {
					return "", err
				}
				prCache[number] = details
			}

			notes, ok := noteCache[number]
			if !ok {
				notes, err = LoadPRReviewNotes(ctx, runner, repo, number, file)
				if err != nil {
					return "", err
				}
				noteCache[number] = notes
			}

			var block strings.Builder
			block.WriteString(header)
			fmt.Fprintf(&block, "PR #%d %s\n", details.Number, details.Title)
			block.WriteString("  why: ")
			block.WriteString(FirstParagraph(details.Body))
			block.WriteString("\n")
			if len(notes) > 0 {
				block.WriteString("  notes: ")
				for i, note := range notes {
					if i > 0 {
						block.WriteString(" | ")
					}
					block.WriteString("@")
					block.WriteString(note.Author)
					block.WriteString(": ")
					block.WriteString(note.Body)
				}
				block.WriteString("\n")
			}
			lines = append(lines, strings.TrimRight(block.String(), "\n"))
			continue
		}

		message, err := GitCommitMessage(ctx, runner, span.SHA)
		if err != nil {
			return "", err
		}

		lines = append(lines, strings.Join([]string{
			header + "direct",
			"  why: " + DirectCommitWhy(message),
		}, "\n"))
	}

	return renderUnearthLines(file, start, end, lines), nil
}

func GitBlameSpans(ctx context.Context, runner execx.Runner, file string, start int, end int) ([]BlameSpan, error) {
	rangeArg := fmt.Sprintf("%d,%d", start, end)
	out, err := RunGit(ctx, runner, "blame", "--line-porcelain", "-L", rangeArg, "--", file)
	if err != nil {
		return nil, err
	}
	return parseBlameSpans(string(out))
}

func LoadPRReviewNotes(ctx context.Context, runner execx.Runner, repo string, number int, file string) ([]ReviewNote, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/comments?per_page=100", repo, number)
	body, err := GHAPI(ctx, runner, "--paginate", "-H", "Accept: application/vnd.github+json", path)
	if err != nil {
		return nil, err
	}

	var payload []reviewCommentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	notes := make([]ReviewNote, 0, len(payload))
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
		notes = append(notes, ReviewNote{
			Author: comment.User.Login,
			Body:   cleanBody,
			When:   comment.CreatedAt,
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

func renderUnearthLines(file string, start int, end int, entries []string) string {
	var out strings.Builder

	fmt.Fprintf(&out, "# %s:%d-%d\n", file, start, end)
	if len(entries) == 0 {
		out.WriteString("\n(none)\n")
		return out.String()
	}

	out.WriteString("\n")
	for i, entry := range entries {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString(entry)
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n")
}

func renderBlameSpan(span BlameSpan) string {
	if span.Start == span.End {
		return fmt.Sprintf("L%d", span.Start)
	}
	return fmt.Sprintf("L%d-L%d", span.Start, span.End)
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
