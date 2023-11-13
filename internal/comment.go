package internal

import (
	"strings"

	"github.com/modernice/jotbot/internal/slice"
)

// RemoveColumns cleans up paragraph text by removing newline characters within
// paragraphs and normalizing the spacing between words. It returns the
// resulting clean text.
func RemoveColumns(str string) string {
	paras := strings.Split(str, "\n\n")
	paras = slice.Map(paras, func(para string) string {
		return strings.ReplaceAll(strings.ReplaceAll(para, "\n", " "), "  ", " ")
	})
	return strings.Join(paras, "\n\n")
}

// Columns breaks a string into lines with a maximum given length, ensuring that
// words are not split across lines and excess whitespace is removed, and
// returns the resulting slice of strings.
func Columns(str string, maxLen int) []string {
	rawLines := strings.Split(str, "\n")
	var lines []string

	for _, rawLine := range rawLines {
		words := strings.Fields(rawLine)
		var line string
		for _, word := range words {
			if len(line)+len(word) >= maxLen {
				line = strings.TrimSpace(line)
				lines = append(lines, line)
				line = ""
			}
			if len(line) > 0 {
				line += " "
			}
			line += word
		}
		lines = append(lines, strings.TrimSpace(line))
	}

	return slice.Map(lines, strings.TrimSpace)
}
