package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/KalleBylin/chester/internal/execx"
)

var prSummaryFields = []string{"number", "title", "body", "url", "mergedAt"}

type PRDetails struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	URL      string `json:"url"`
	MergedAt string `json:"mergedAt"`
}

type commitPullRef struct {
	Number   int    `json:"number"`
	MergedAt string `json:"merged_at"`
}

type CommitPRRef struct {
	Number int
	Source string
}

func ResolveCommitPRNumberViaGH(ctx context.Context, runner execx.Runner, repo string, sha string) (int, bool, error) {
	path := fmt.Sprintf("repos/%s/commits/%s/pulls", repo, sha)
	body, err := GHAPI(ctx, runner, "-H", "Accept: application/vnd.github+json", path)
	if err != nil {
		return 0, false, err
	}

	var refs []commitPullRef
	if err := json.Unmarshal(body, &refs); err != nil {
		return 0, false, err
	}
	for _, ref := range refs {
		if ref.MergedAt != "" {
			return ref.Number, true, nil
		}
	}
	return 0, false, nil
}

func ResolveCommitPRNumber(ctx context.Context, runner execx.Runner, repo string, sha string, fallbackSubject string) (int, bool, error) {
	subject := strings.TrimSpace(fallbackSubject)
	var err error
	if subject == "" {
		subject, err = GitCommitSubject(ctx, runner, sha)
		if err != nil {
			return 0, false, err
		}
	}

	if number, ok := ExtractPRNumberFromSubject(subject); ok {
		return number, true, nil
	}
	if repo == "" {
		return 0, false, nil
	}
	return ResolveCommitPRNumberViaGH(ctx, runner, repo, sha)
}

func InferCommitPRRef(ctx context.Context, runner execx.Runner, repo string, sha string, fallbackSubject string) (CommitPRRef, bool, error) {
	subject := strings.TrimSpace(fallbackSubject)
	var err error
	if subject == "" {
		subject, err = GitCommitSubject(ctx, runner, sha)
		if err != nil {
			return CommitPRRef{}, false, err
		}
	}

	if number, ok := ExtractPRNumberFromSubject(subject); ok {
		return CommitPRRef{
			Number: number,
			Source: "commit_subject",
		}, true, nil
	}

	if repo == "" {
		return CommitPRRef{}, false, nil
	}

	number, ok, err := ResolveCommitPRNumberViaGH(ctx, runner, repo, sha)
	if err != nil {
		return CommitPRRef{}, false, nil
	}
	if !ok {
		return CommitPRRef{}, false, nil
	}
	return CommitPRRef{
		Number: number,
		Source: "github_api",
	}, true, nil
}

func ExtractPRNumberFromSubject(subject string) (int, bool) {
	const mergePrefix = "Merge pull request #"
	if strings.HasPrefix(subject, mergePrefix) {
		digits := readLeadingDigits(subject[len(mergePrefix):])
		if digits != "" {
			value, _ := strconv.Atoi(digits)
			return value, true
		}
	}

	start := strings.LastIndex(subject, "(#")
	if start >= 0 && strings.HasSuffix(subject, ")") {
		digits := subject[start+2 : len(subject)-1]
		if digits != "" && digits == readLeadingDigits(digits) {
			value, _ := strconv.Atoi(digits)
			return value, true
		}
	}

	return 0, false
}

func LoadPRDetails(ctx context.Context, runner execx.Runner, repo string, number int) (PRDetails, error) {
	body, err := GHPRView(ctx, runner, repo, strconv.Itoa(number), prSummaryFields)
	if err != nil {
		return PRDetails{}, err
	}

	var details PRDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return PRDetails{}, err
	}
	return details, nil
}

func TryLoadPRDetails(ctx context.Context, runner execx.Runner, repo string, number int) (PRDetails, bool) {
	if repo == "" {
		return PRDetails{}, false
	}

	details, err := LoadPRDetails(ctx, runner, repo, number)
	if err != nil {
		return PRDetails{}, false
	}
	return details, true
}

func PRSummary(details PRDetails) Summary {
	if summary := FirstParagraph(details.Body); summary != "" {
		return Summary{
			Text:   summary,
			Source: "pr_body",
		}
	}
	return Summary{
		Text:   details.Title,
		Source: "pr_title",
	}
}

func CommitSummary(raw string) Summary {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Summary{}
	}

	parts := strings.SplitN(raw, "\n\n", 2)
	subject := strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		if paragraph := FirstParagraph(parts[1]); paragraph != "" {
			return Summary{
				Text:   paragraph,
				Source: "commit_body",
			}
		}
	}
	return Summary{
		Text:   subject,
		Source: "commit_subject",
	}
}

func shortSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}

func readLeadingDigits(value string) string {
	var out strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			break
		}
		out.WriteByte(value[i])
	}
	return out.String()
}

func NewCommitRef(sha string) CommitRef {
	return CommitRef{
		SHA:   sha,
		Short: shortSHA(sha),
	}
}
