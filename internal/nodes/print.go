package nodes

import (
	"bytes"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func Fprint(w io.Writer, node *dst.File) error {
	return decorator.Fprint(w, node)
}

func Format(node *dst.File) ([]byte, error) {
	var buf bytes.Buffer
	err := decorator.Fprint(&buf, node)
	return buf.Bytes(), err
}
