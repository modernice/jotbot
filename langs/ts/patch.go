package ts

import (
	"fmt"
	"strings"
	"unicode"
)

// InsertComment inserts a given comment into a slice of bytes representing code
// at the specified position. It returns the modified code with the comment
// inserted and an error if the position is out of range. The comment is
// inserted in a way that aligns with the indentation of the line at the given
// position. If successful, it returns the updated code as a []byte and nil
// error; otherwise, it returns nil and an error describing the invalid
// position.
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
