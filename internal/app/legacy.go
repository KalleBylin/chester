package app

import (
	"context"

	"github.com/KalleBylin/chester/internal/execx"
)

func FileHistory(ctx context.Context, runner execx.Runner, repo string, path string) (string, error) {
	result, err := WhyFile(ctx, runner, repo, path)
	if err != nil {
		return "", err
	}
	return RenderWhyFileMarkdown(result), nil
}

func UnearthRange(ctx context.Context, runner execx.Runner, repo string, spec string) (string, error) {
	result, err := WhyRange(ctx, runner, repo, spec)
	if err != nil {
		return "", err
	}
	return RenderWhyRangeMarkdown(result), nil
}

func UnearthLines(ctx context.Context, runner execx.Runner, repo string, file string, start int, end int) (string, error) {
	result, err := WhyLines(ctx, runner, repo, file, start, end)
	if err != nil {
		return "", err
	}
	return RenderWhyLinesMarkdown(result), nil
}

func LoadPRReviewNotes(ctx context.Context, runner execx.Runner, repo string, number int, file string) ([]ReviewComment, error) {
	return LoadPRReviewComments(ctx, runner, repo, number, file)
}

func PRWhy(details PRDetails) string {
	return PRSummary(details).Text
}

func DirectCommitWhy(raw string) string {
	return CommitSummary(raw).Text
}
