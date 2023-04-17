package git

import "strings"

type Commit struct {
	Msg    string
	Desc   []string
	Footer string
}

func NewCommit(msg string, desc ...string) Commit {
	return Commit{
		Msg:  msg,
		Desc: desc,
	}
}

func DefaultCommit() Commit {
	c := NewCommit("docs: add missing documentation")
	c.Footer = "This commit was created by opendocs."
	return c
}

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

func (c Commit) String() string {
	return strings.Join(c.Paragraphs(), "\n\n")
}
