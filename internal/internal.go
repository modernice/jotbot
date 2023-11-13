package internal

import "strings"

// JoinStrings concatenates the elements of a slice of strings, using the
// provided separator between adjacent elements, and returns the resulting
// string. If the slice is empty, it returns an empty string. If the slice has
// only one element, it returns that element without the separator.
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
