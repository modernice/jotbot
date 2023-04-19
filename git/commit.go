package git

import "strings"

// Commit represents a Git commit. It contains a message, a description, and a 
// footer. The message is a short summary of the changes made in the commit. The 
// description is an optional longer explanation of the changes. The footer is 
// an optional section for metadata or references. The Commit type has methods 
// for creating new commits, checking equality with other commits, and 
// formatting the commit as a string.
type Commit struct {
	Msg    string
	Desc   []string
	Footer string
}

// NewCommit creates a new Commit with the given message and description. It 
// returns a Commit struct. The message is a required parameter, while the 
// description is optional and can be passed as a variadic argument.
func NewCommit(msg string, desc ...string) Commit {
	return Commit{
		Msg:  msg,
		Desc: desc,
	}
}

// DefaultCommit returns a Commit with a default commit message and footer. The 
// commit message is "docs: add missing documentation" and the footer is "This 
// commit was created by opendocs."
func DefaultCommit() Commit {
	c := NewCommit("docs: add missing documentation")
	c.Footer = "This commit was created by opendocs."
	return c
}

// Equal determines whether two Commit values are equal. It returns true if the 
// Msg and Footer fields of the two Commits are equal, and if their Desc fields 
// are equal as determined by the allEqual function.
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

// Paragraphs returns a slice of strings containing the commit message, 
// description, and footer of a [Commit] object, separated by newlines.
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

// String returns a string representation of the Commit, including the commit 
// message, description, and footer, separated by two newlines.
func (c Commit) String() string {
	return strings.Join(c.Paragraphs(), "\n\n")
}
