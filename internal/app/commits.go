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

func ResolveCommitPRNumber(ctx context.Context, runner execx.Runner, repo string, sha string, fallbackSubject string) (int, bool, error) {
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

	subject := fallbackSubject
	if subject == "" {
		subject, err = GitCommitSubject(ctx, runner, sha)
		if err != nil {
			return 0, false, err
		}
	}

	number, ok := ExtractPRNumberFromSubject(subject)
	if !ok {
		return 0, false, nil
	}
	return number, true, nil
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

func PRWhy(details PRDetails) string {
	if why := FirstParagraph(details.Body); why != "" {
		return why
	}
	return details.Title
}

func DirectCommitWhy(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parts := strings.SplitN(raw, "\n\n", 2)
	subject := strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		if paragraph := FirstParagraph(parts[1]); paragraph != "" {
			return paragraph
		}
	}
	return subject
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
