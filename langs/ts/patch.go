package ts

import (
	"fmt"
	"strings"
	"unicode"
)

func InsertComment(comment string, code []byte, pos Position) ([]byte, error) {
	lines := strings.Split(string(code), "\n")
	commentLines := strings.Split(comment, "\n")

	if pos.Line >= len(lines) || pos.Line < 0 {
		return nil, fmt.Errorf("line number %d out of range", pos.Line)
	}

	targetLine := lines[pos.Line]
	if pos.Character > len(targetLine) || pos.Character < 0 {
		return nil, fmt.Errorf("character position %d out of range", pos.Character)
	}

	prefix := ""
	for _, r := range targetLine {
		if unicode.IsSpace(r) {
			prefix += string(r)
			continue
		}
		break
	}

	for i, line := range commentLines[1:] {
		commentLines[i+1] = prefix + line
	}

	comment = strings.Join(commentLines, "\n")

	modifiedLine := targetLine[:pos.Character] + comment + targetLine[pos.Character:]
	lines[pos.Line] = modifiedLine

	return []byte(strings.Join(lines, "\n")), nil
}
