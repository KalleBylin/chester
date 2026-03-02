package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"chester/internal/execx"
)

var (
	readThreadPRFields    = []string{"number", "title", "body", "url", "state", "isDraft", "comments", "reviews"}
	readThreadIssueFields = []string{"number", "title", "body", "url", "state", "comments"}
)

type readThreadPRPayload struct {
	Number   int                 `json:"number"`
	Title    string              `json:"title"`
	Body     string              `json:"body"`
	URL      string              `json:"url"`
	State    string              `json:"state"`
	IsDraft  bool                `json:"isDraft"`
	Comments []readThreadComment `json:"comments"`
	Reviews  []readThreadReview  `json:"reviews"`
}

type readThreadIssuePayload struct {
	Number   int                 `json:"number"`
	Title    string              `json:"title"`
	Body     string              `json:"body"`
	URL      string              `json:"url"`
	State    string              `json:"state"`
	Comments []readThreadComment `json:"comments"`
}

type readThreadComment struct {
	Author    readThreadAuthor `json:"author"`
	Body      string           `json:"body"`
	CreatedAt string           `json:"createdAt"`
}

type readThreadReview struct {
	Author      readThreadAuthor `json:"author"`
	Body        string           `json:"body"`
	State       string           `json:"state"`
	SubmittedAt string           `json:"submittedAt"`
}

type readThreadAuthor struct {
	Login string `json:"login"`
	IsBot bool   `json:"isBot"`
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
		return renderPRThread(body)
	}
	if !isPRLookupMiss(err) {
		return "", err
	}

	body, err = GHIssueView(ctx, runner, repo, id, readThreadIssueFields)
	if err != nil {
		return "", err
	}
	return renderIssueThread(body)
}

func renderPRThread(body []byte) (string, error) {
	var payload readThreadPRPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	events := make([]readThreadEvent, 0, len(payload.Comments)+len(payload.Reviews))
	for _, comment := range payload.Comments {
		rendered, ok := newCommentEvent(comment)
		if ok {
			events = append(events, rendered)
		}
	}
	for _, review := range payload.Reviews {
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

func renderIssueThread(body []byte) (string, error) {
	var payload readThreadIssuePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	events := make([]readThreadEvent, 0, len(payload.Comments))
	for _, comment := range payload.Comments {
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

func newCommentEvent(comment readThreadComment) (readThreadEvent, bool) {
	if comment.Author.IsBot {
		return readThreadEvent{}, false
	}

	body := SanitizeMarkdown(comment.Body)
	if body == "" {
		return readThreadEvent{}, false
	}

	return readThreadEvent{
		When:  comment.CreatedAt,
		Label: "@" + comment.Author.Login,
		Body:  body,
	}, true
}

func newReviewEvent(review readThreadReview) (readThreadEvent, bool) {
	if review.Author.IsBot {
		return readThreadEvent{}, false
	}

	body := SanitizeMarkdown(review.Body)
	if body == "" {
		return readThreadEvent{}, false
	}

	return readThreadEvent{
		When:  review.SubmittedAt,
		Label: "review/" + review.State + " @" + review.Author.Login,
		Body:  body,
	}, true
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
