package app

import "strings"

func SanitizeMarkdown(input string) string {
	cleaned := stripHTMLComments(input)
	lines := strings.Split(cleaned, "\n")

	filtered := make([]string, 0, len(lines))
	blankPending := false
	for _, line := range lines {
		if isEmptyCheckboxLine(line) {
			continue
		}

		if strings.TrimSpace(line) == "" {
			if blankPending {
				continue
			}
			filtered = append(filtered, "")
			blankPending = true
			continue
		}

		filtered = append(filtered, strings.TrimRight(line, " \t"))
		blankPending = false
	}

	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func FirstParagraph(input string) string {
	cleaned := SanitizeMarkdown(input)
	if cleaned == "" {
		return ""
	}

	paragraphs := strings.Split(cleaned, "\n\n")
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		lines := strings.Split(paragraph, "\n")
		parts := make([]string, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts = append(parts, line)
		}
		return strings.Join(parts, " ")
	}

	return ""
}

func stripHTMLComments(input string) string {
	var out strings.Builder
	inComment := false

	for i := 0; i < len(input); {
		switch {
		case !inComment && strings.HasPrefix(input[i:], "<!--"):
			inComment = true
			i += len("<!--")
		case inComment && strings.HasPrefix(input[i:], "-->"):
			inComment = false
			i += len("-->")
		case inComment:
			i++
		default:
			out.WriteByte(input[i])
			i++
		}
	}

	return out.String()
}

func isEmptyCheckboxLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	prefixes := []string{"- [ ]", "- [x]", "- [X]", "* [ ]", "* [x]", "* [X]"}

	for _, prefix := range prefixes {
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		return rest == ""
	}
	return false
}
