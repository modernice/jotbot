package nodes

import (
	"bytes"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// Fprint formats and writes the source code of a given *dst.File to an
// io.Writer, using Go syntax. It is a convenience wrapper around
// decorator.Fprint.
func Fprint(w io.Writer, node *dst.File) error {
	return decorator.Fprint(w, node)
}

// Format returns the formatted source code of a given *dst.File node as a
// []byte slice, and an error if one occurred during formatting.
func Format(node *dst.File) ([]byte, error) {
	var buf bytes.Buffer
	err := decorator.Fprint(&buf, node)
	return buf.Bytes(), err
}
