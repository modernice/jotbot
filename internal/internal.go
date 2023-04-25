package internal

import "strings"

// Join concatenates the elements of its first argument to create a single string. The separator
// string sep is placed between elements in the resulting string.
//
// Copied and adapted from https://go.dev/src/strings/strings.go
func JoinStrings[E ~string](elems []E, sep string) E {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(string(elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(string(s))
	}
	return E(b.String())
}
