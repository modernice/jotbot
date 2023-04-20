package nodes

import (
	"bytes"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// Fprint formats a *dst.File and writes the result to an io.Writer. It uses the
// decorator.Fprint function from the "github.com/dave/dst/decorator" package to
// format the AST nodes.
func Fprint(w io.Writer, node *dst.File) error {
	return decorator.Fprint(w, node)
}

// Format formats a Go AST [dst.File] into a []byte. It uses the decorator
// package to print the AST with additional formatting such as indentation and
// comments.
func Format(node *dst.File) ([]byte, error) {
	var buf bytes.Buffer
	err := decorator.Fprint(&buf, node)
	return buf.Bytes(), err
}
