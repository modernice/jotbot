package git

import "strings"

// Commit represents a set of changes or updates in a version control system
// with an associated message, optional extended description, and an optional
// footer. It provides the ability to compare itself with another commit for
// equality, generate a string representation consisting of its message,
// description paragraphs, and footer, and create default or custom commits with
// the provided message and description.
type Commit struct {
	Msg    string
	Desc   []string
	Footer string
}

// NewCommit creates a new Commit with a message and an optional description. It
// returns the created Commit.
func NewCommit(msg string, desc ...string) Commit {
	return Commit{
		Msg:  msg,
		Desc: desc,
	}
}

// DefaultCommit creates a new Commit with a predefined message and footer
// indicating that the commit was created automatically.
func DefaultCommit() Commit {
	c := NewCommit("docs: add missing documentation")
	c.Footer = "This commit was created by jotbot."
	return c
}

// Equal reports whether two [Commit] instances are considered equivalent,
// comparing the message, footer, and description slices.
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

// Paragraphs constructs a slice of strings representing the structured content
// of a commit, including its message, optional description paragraphs, and
// footer if present. Each element in the slice corresponds to a distinct
// section of the commit content. The message is always included as the first
// element, followed by the joined description lines as a single element if they
// exist, and finally the footer as the last element if it is not empty. The
// function ensures that even if the commit message is initially empty, a
// default message is provided. It returns a slice of strings where each string
// represents a separate paragraph or section of the commit.
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

// String returns a string representation of the commit, combining the message,
// description, and footer with appropriate spacing.
func (c Commit) String() string {
	return strings.Join(c.Paragraphs(), "\n\n")
}
