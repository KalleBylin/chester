package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/KalleBylin/chester/internal/execx"
)

var (
	readThreadPRFields    = []string{"number", "title", "body", "url", "state", "isDraft"}
	readThreadIssueFields = []string{"number", "title", "body", "url", "state"}
)

type readThreadPRPayload struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	URL     string `json:"url"`
	State   string `json:"state"`
	IsDraft bool   `json:"isDraft"`
}

type readThreadIssuePayload struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	URL    string `json:"url"`
	State  string `json:"state"`
}

type readThreadIssueComment struct {
	User struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type readThreadReview struct {
	User struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
	Body        string `json:"body"`
	State       string `json:"state"`
	SubmittedAt string `json:"submitted_at"`
}

type readThreadEvent struct {
	When   string
	Label  string
	Body   string
	Review bool
}

func ReadThread(ctx context.Context, runner execx.Runner, repo string, id string) (string, error) {
	body, err := GHPRView(ctx, runner, repo, id, readThreadPRFields)
	if err == nil {
		return renderPRThread(ctx, runner, repo, body)
	}
	if !isPRLookupMiss(err) {
		return "", err
	}

	body, err = GHIssueView(ctx, runner, repo, id, readThreadIssueFields)
	if err != nil {
		return "", err
	}
	return renderIssueThread(ctx, runner, repo, body)
}

func renderPRThread(ctx context.Context, runner execx.Runner, repo string, body []byte) (string, error) {
	var payload readThreadPRPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	comments, err := loadIssueComments(ctx, runner, repo, payload.Number)
	if err != nil {
		return "", err
	}
	reviews, err := loadPRReviews(ctx, runner, repo, payload.Number)
	if err != nil {
		return "", err
	}

	events := make([]readThreadEvent, 0, len(comments)+len(reviews))
	for _, comment := range comments {
		rendered, ok := newCommentEvent(comment)
		if ok {
			events = append(events, rendered)
		}
	}
	for _, review := range reviews {
		rendered, ok := newReviewEvent(review)
		if ok {
			events = append(events, rendered)
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].When < events[j].When
	})

	return renderThreadDocument(payload.Number, "pr", payload.Title, payload.URL, payload.Body, events), nil
}

func renderIssueThread(ctx context.Context, runner execx.Runner, repo string, body []byte) (string, error) {
	var payload readThreadIssuePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	comments, err := loadIssueComments(ctx, runner, repo, payload.Number)
	if err != nil {
		return "", err
	}

	events := make([]readThreadEvent, 0, len(comments))
	for _, comment := range comments {
		rendered, ok := newCommentEvent(comment)
		if ok {
			events = append(events, rendered)
		}
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].When < events[j].When
	})

	return renderThreadDocument(payload.Number, "issue", payload.Title, payload.URL, payload.Body, events), nil
}

func newCommentEvent(comment readThreadIssueComment) (readThreadEvent, bool) {
	if comment.User.Type != "User" {
		return readThreadEvent{}, false
	}

	body := SanitizeMarkdown(comment.Body)
	if body == "" {
		return readThreadEvent{}, false
	}

	return readThreadEvent{
		When:  comment.CreatedAt,
		Label: "@" + comment.User.Login,
		Body:  body,
	}, true
}

func newReviewEvent(review readThreadReview) (readThreadEvent, bool) {
	if review.User.Type != "User" {
		return readThreadEvent{}, false
	}

	body := SanitizeMarkdown(review.Body)
	if body == "" {
		return readThreadEvent{}, false
	}

	return readThreadEvent{
		When:  review.SubmittedAt,
		Label: "review/" + review.State + " @" + review.User.Login,
		Body:  body,
	}, true
}

func loadIssueComments(ctx context.Context, runner execx.Runner, repo string, number int) ([]readThreadIssueComment, error) {
	path := fmt.Sprintf("repos/%s/issues/%d/comments?per_page=100", repo, number)
	body, err := GHAPI(ctx, runner, "--paginate", "-H", "Accept: application/vnd.github+json", path)
	if err != nil {
		return nil, err
	}

	var comments []readThreadIssueComment
	if err := json.Unmarshal(body, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func loadPRReviews(ctx context.Context, runner execx.Runner, repo string, number int) ([]readThreadReview, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/reviews?per_page=100", repo, number)
	body, err := GHAPI(ctx, runner, "--paginate", "-H", "Accept: application/vnd.github+json", path)
	if err != nil {
		return nil, err
	}

	var reviews []readThreadReview
	if err := json.Unmarshal(body, &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func renderThreadDocument(number int, kind string, title string, url string, body string, events []readThreadEvent) string {
	var out strings.Builder

	fmt.Fprintf(&out, "#%d [%s] %s\n", number, kind, title)
	out.WriteString(url)
	out.WriteString("\n\n## Body\n")

	cleanBody := SanitizeMarkdown(body)
	if cleanBody == "" {
		out.WriteString("(empty)\n")
	} else {
		out.WriteString(cleanBody)
		out.WriteString("\n")
	}

	out.WriteString("\n## Thread\n")
	if len(events) == 0 {
		out.WriteString("(none)\n")
		return out.String()
	}

	for i, event := range events {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("- ")
		out.WriteString(event.Label)
		out.WriteString("\n")
		writeIndentedBlock(&out, event.Body)
	}

	return out.String()
}

func writeIndentedBlock(out *strings.Builder, body string) {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("  ")
		out.WriteString(line)
	}
	out.WriteString("\n")
}

func isPRLookupMiss(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrGHNotInstalled) || errors.Is(err, ErrGHUnauthenticated) {
		return false
	}

	var runErr *execx.RunError
	if !errors.As(err, &runErr) {
		return false
	}

	stderr := strings.ToLower(strings.TrimSpace(string(runErr.Stderr)))
	fragments := []string{
		"no pull request found",
		"no pull requests found",
		"could not resolve to a pull request",
	}
	for _, fragment := range fragments {
		if strings.Contains(stderr, fragment) {
			return true
		}
	}
	return false
}
