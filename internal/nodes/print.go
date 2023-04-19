package nodes

import (
	"bytes"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// Fprint writes the formatted source code of a given *dst.File node to an
// io.Writer, using decorator.Fprint to perform the formatting.
func Fprint(w io.Writer, node *dst.File) error {
	return decorator.Fprint(w, node)
}

// Format returns the formatted source code of a given *dst.File node as a
// []byte. It uses decorator.Fprint to format the node and returns any errors
// encountered during the formatting process.
func Format(node *dst.File) ([]byte, error) {
	var buf bytes.Buffer
	err := decorator.Fprint(&buf, node)
	return buf.Bytes(), err
}
