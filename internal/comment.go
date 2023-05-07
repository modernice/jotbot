package internal

import (
	"strings"

	"github.com/modernice/jotbot/internal/slice"
)

// Columns splits the given string into a slice of strings, each with a maximum
// length of maxLen, preserving the original words and line breaks.
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
