package app

import "testing"

func TestSanitizeMarkdownRemovesNoise(t *testing.T) {
	t.Parallel()

	input := "First paragraph.\n\n<!-- hidden -->\n\n- [ ]\n\nSecond paragraph.\n\n\n"
	got := SanitizeMarkdown(input)
	want := "First paragraph.\n\nSecond paragraph."
	if got != want {
		t.Fatalf("SanitizeMarkdown() = %q, want %q", got, want)
	}
}

func TestFirstParagraphReturnsFirstNonEmptyParagraph(t *testing.T) {
	t.Parallel()

	input := "\n\n<!-- hidden -->\nLine one\nLine two\n\nSecond block"
	got := FirstParagraph(input)
	want := "Line one Line two"
	if got != want {
		t.Fatalf("FirstParagraph() = %q, want %q", got, want)
	}
}

func TestFirstParagraphReturnsEmptyForEmptyInput(t *testing.T) {
	t.Parallel()

	if got := FirstParagraph("<!-- hidden -->\n- [ ]\n"); got != "" {
		t.Fatalf("FirstParagraph() = %q, want empty string", got)
	}
}
