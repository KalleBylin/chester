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

func ReadThread(ctx context.Context, runner execx.Runner, repo string, id string) (ReadThreadResult, error) {
	body, err := GHPRView(ctx, runner, repo, id, readThreadPRFields)
	if err == nil {
		return renderPRThread(ctx, runner, repo, body)
	}
	if !isPRLookupMiss(err) {
		return ReadThreadResult{}, err
	}

	body, err = GHIssueView(ctx, runner, repo, id, readThreadIssueFields)
	if err != nil {
		return ReadThreadResult{}, err
	}
	return renderIssueThread(ctx, runner, repo, body)
}

func renderPRThread(ctx context.Context, runner execx.Runner, repo string, body []byte) (ReadThreadResult, error) {
	var payload readThreadPRPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return ReadThreadResult{}, err
	}

	comments, err := loadIssueComments(ctx, runner, repo, payload.Number)
	if err != nil {
		return ReadThreadResult{}, err
	}
	reviews, err := loadPRReviews(ctx, runner, repo, payload.Number)
	if err != nil {
		return ReadThreadResult{}, err
	}
	reviewComments, err := loadPRInlineReviewComments(ctx, runner, repo, payload.Number)
	if err != nil {
		return ReadThreadResult{}, err
	}

	threadComments := make([]ThreadComment, 0, len(comments))
	for _, comment := range comments {
		rendered, ok := newThreadComment(comment)
		if ok {
			threadComments = append(threadComments, rendered)
		}
	}
	sort.SliceStable(threadComments, func(i, j int) bool {
		return threadComments[i].When < threadComments[j].When
	})

	threadReviews := make([]ThreadReview, 0, len(reviews))
	for _, review := range reviews {
		rendered, ok := newThreadReview(review)
		if ok {
			threadReviews = append(threadReviews, rendered)
		}
	}
	sort.SliceStable(threadReviews, func(i, j int) bool {
		return threadReviews[i].When < threadReviews[j].When
	})

	return ReadThreadResult{
		Number:         payload.Number,
		Kind:           "pr",
		Title:          payload.Title,
		URL:            payload.URL,
		Body:           SanitizeMarkdown(payload.Body),
		Comments:       threadComments,
		Reviews:        threadReviews,
		ReviewComments: reviewComments,
	}, nil
}

func renderIssueThread(ctx context.Context, runner execx.Runner, repo string, body []byte) (ReadThreadResult, error) {
	var payload readThreadIssuePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return ReadThreadResult{}, err
	}

	comments, err := loadIssueComments(ctx, runner, repo, payload.Number)
	if err != nil {
		return ReadThreadResult{}, err
	}

	events := make([]ThreadComment, 0, len(comments))
	for _, comment := range comments {
		rendered, ok := newThreadComment(comment)
		if ok {
			events = append(events, rendered)
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].When < events[j].When
	})

	return ReadThreadResult{
		Number:   payload.Number,
		Kind:     "issue",
		Title:    payload.Title,
		URL:      payload.URL,
		Body:     SanitizeMarkdown(payload.Body),
		Comments: events,
	}, nil
}

func newThreadComment(comment readThreadIssueComment) (ThreadComment, bool) {
	if comment.User.Type != "User" {
		return ThreadComment{}, false
	}

	body := SanitizeMarkdown(comment.Body)
	if body == "" {
		return ThreadComment{}, false
	}

	return ThreadComment{
		When:   comment.CreatedAt,
		Author: comment.User.Login,
		Body:   body,
	}, true
}

func newThreadReview(review readThreadReview) (ThreadReview, bool) {
	if review.User.Type != "User" {
		return ThreadReview{}, false
	}

	body := SanitizeMarkdown(review.Body)
	if body == "" {
		return ThreadReview{}, false
	}

	return ThreadReview{
		When:   review.SubmittedAt,
		Author: review.User.Login,
		State:  review.State,
		Body:   body,
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

func loadPRInlineReviewComments(ctx context.Context, runner execx.Runner, repo string, number int) ([]ReviewComment, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/comments?per_page=100", repo, number)
	body, err := GHAPI(ctx, runner, "--paginate", "-H", "Accept: application/vnd.github+json", path)
	if err != nil {
		return nil, err
	}

	var payload []reviewCommentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	comments := make([]ReviewComment, 0, len(payload))
	for _, comment := range payload {
		if comment.User.Type != "User" {
			continue
		}
		cleanBody := SanitizeMarkdown(comment.Body)
		if cleanBody == "" {
			continue
		}
		startLine, endLine := commentRange(comment)
		comments = append(comments, ReviewComment{
			Author:    comment.User.Login,
			Body:      cleanBody,
			When:      comment.CreatedAt,
			Path:      comment.Path,
			StartLine: startLine,
			EndLine:   endLine,
		})
	}

	sort.SliceStable(comments, func(i, j int) bool {
		return comments[i].When < comments[j].When
	})
	return comments, nil
}

func RenderReadThreadMarkdown(result ReadThreadResult) string {
	var out strings.Builder

	fmt.Fprintf(&out, "#%d [%s] %s\n", result.Number, result.Kind, result.Title)
	out.WriteString(result.URL)
	out.WriteString("\n\n## Body\n")

	if result.Body == "" {
		out.WriteString("(empty)\n")
	} else {
		out.WriteString(result.Body)
		out.WriteString("\n")
	}

	renderCommentSection(&out, "Comments", result.Comments)
	if result.Kind == "pr" {
		renderReviewSection(&out, result.Reviews)
		renderReviewCommentSection(&out, result.ReviewComments)
	}
	return out.String()
}

func renderCommentSection(out *strings.Builder, heading string, comments []ThreadComment) {
	fmt.Fprintf(out, "\n## %s\n", heading)
	if len(comments) == 0 {
		out.WriteString("(none)\n")
		return
	}
	for i, comment := range comments {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("- @")
		out.WriteString(comment.Author)
		out.WriteString("\n")
		writeIndentedBlock(out, comment.Body)
	}
}

func renderReviewSection(out *strings.Builder, reviews []ThreadReview) {
	out.WriteString("\n## Reviews\n")
	if len(reviews) == 0 {
		out.WriteString("(none)\n")
		return
	}
	for i, review := range reviews {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("- ")
		out.WriteString(review.State)
		out.WriteString(" @")
		out.WriteString(review.Author)
		out.WriteString("\n")
		writeIndentedBlock(out, review.Body)
	}
}

func renderReviewCommentSection(out *strings.Builder, comments []ReviewComment) {
	out.WriteString("\n## Review Comments\n")
	if len(comments) == 0 {
		out.WriteString("(none)\n")
		return
	}
	for i, comment := range comments {
		if i > 0 {
			out.WriteString("\n")
		}
		out.WriteString("- ")
		out.WriteString(renderInlineReviewLabel(comment))
		out.WriteString("\n")
		writeIndentedBlock(out, comment.Body)
	}
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

func renderInlineReviewLabel(comment ReviewComment) string {
	label := "@" + comment.Author
	if comment.Path != "" {
		label = comment.Path + " " + label
	}
	switch {
	case comment.StartLine > 0 && comment.EndLine > 0 && comment.StartLine != comment.EndLine:
		return fmt.Sprintf("%s (L%d-L%d)", label, comment.StartLine, comment.EndLine)
	case comment.StartLine > 0:
		return fmt.Sprintf("%s (L%d)", label, comment.StartLine)
	default:
		return label
	}
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
