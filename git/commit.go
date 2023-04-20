package git

import "strings"

// Commit represents a single commit in a Git repository. It contains a commit
// message (Msg), a description (Desc), and a footer (Footer). NewCommit creates
// a new commit with the given message and description, while DefaultCommit
// creates a new commit with default values for the message and footer. The
// Equal method determines whether two commits are equal by comparing their
// message, footer, and description. The Paragraphs method returns the commit's
// message, description, and footer as separate paragraphs. The String method
// returns the commit as a single string with paragraphs separated by two
// newlines.
type Commit struct {
	Msg    string
	Desc   []string
	Footer string
}

// NewCommit creates a new Git commit with the given commit message and optional
// description. It returns a Commit type.
func NewCommit(msg string, desc ...string) Commit {
	return Commit{
		Msg:  msg,
		Desc: desc,
	}
}

// DefaultCommit creates a Commit with a default commit message and footer. The
// commit message is "docs: add missing documentation" and the footer is "This
// commit was created by jotbot."
func DefaultCommit() Commit {
	c := NewCommit("docs: add missing documentation")
	c.Footer = "This commit was created by jotbot."
	return c
}

// Equal checks if two "Commit" values are equal. It returns true if the "Msg",
// "Footer", and "Desc" fields of both values are equal.
func (c Commit) Equal(c2 Commit) bool {
	return c.Msg == c2.Msg && c.Footer == c2.Footer && allEqual(c.Desc, c2.Desc)
}

func allEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Paragraphs returns a slice of strings representing the commit message,
// description, and footer of a Commit struct.
func (c Commit) Paragraphs() []string {
	out := make([]string, 0, len(c.Desc)+2)
	if c.Msg == "" {
		c.Msg = "docs: add missing documentation"
	}
	out = append(out, c.Msg)
	if len(c.Desc) > 0 {
		out = append(out, strings.Join(c.Desc, "\n"))
	}
	if c.Footer != "" {
		out = append(out, c.Footer)
	}
	return out
}

// String returns a string representation of the Commit object, formatted as a
// series of paragraphs separated by two newline characters.
func (c Commit) String() string {
	return strings.Join(c.Paragraphs(), "\n\n")
}
