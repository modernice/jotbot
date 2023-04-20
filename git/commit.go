package git

import "strings"

// Commit represents a Git commit message. It consists of a message, an optional 
// description, and an optional footer. The Commit type provides methods to 
// create new commits, check for equality with other commits, and format the 
// commit as a string. [NewCommit], [DefaultCommit], [Equal], [Paragraphs], and 
// [String] are the relevant functions for working with Commit objects.
type Commit struct {
	Msg    string
	Desc   []string
	Footer string
}

// NewCommit [func] creates a new Commit object with the specified message and 
// description. The function takes a string msg and a variadic slice of strings 
// desc, and returns a Commit object. If no description is given, an empty slice 
// is used instead.
func NewCommit(msg string, desc ...string) Commit {
	return Commit{
		Msg:  msg,
		Desc: desc,
	}
}

// DefaultCommit returns a new [Commit](#Commit) with a default commit message 
// and footer. The message will be "docs: add missing documentation" and the 
// footer will be "This commit was created by jotbot."
func DefaultCommit() Commit {
	c := NewCommit("docs: add missing documentation")
	c.Footer = "This commit was created by jotbot."
	return c
}

// Equal determines whether two Commits are equal. Two Commits are considered 
// equal if they have the same message, footer, and description(s) [Commit, 
// strings].
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

// Paragraphs returns a slice of strings representing the paragraphs of the 
// commit message. The first string is the commit message itself, followed by 
// any description lines, and ending with the footer line.
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

// String returns the commit message, description, and footer as a single string 
// with double line breaks between each paragraph. It is a method of the Commit 
// type [Commit].
func (c Commit) String() string {
	return strings.Join(c.Paragraphs(), "\n\n")
}
