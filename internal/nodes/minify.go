package nodes

import (
	"github.com/dave/dst"
)

func Minify[Node dst.Node](in Node) Node {
	out := dst.Clone(in)
	if _, ok := out.(Node); !ok {
		panic("dst.Clone() returned a wrong type. this should be impossible")
	}

	dst.Inspect(out, func(node dst.Node) bool {
		if _, exported := Identifier(node); exported {
			return true
		}

		switch node := node.(type) {
		case *dst.FuncDecl:
			node.Body = nil
		}

		return true
	})

	return out.(Node)
}
