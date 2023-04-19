package nodes

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func MinifyCode(code []byte) (*dst.File, error) {
	node, err := decorator.Parse(code)
	if err != nil {
		return nil, err
	}
	return Minify(node), nil
}

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
			node.Decs.Start = nil
		}

		return true
	})

	return out.(Node)
}
