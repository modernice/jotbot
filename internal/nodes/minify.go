package nodes

import (
	"github.com/dave/dst"
)

type MinifyOptions struct {
	PackageComment bool
	FuncComment    bool
	FuncBody       bool
	Exported       bool
}

func (opts MinifyOptions) Minify(node dst.Node) dst.Node {
	out := dst.Clone(node)

	patch := func(node dst.Node) {
		switch node := node.(type) {
		case *dst.FuncDecl:
			if opts.FuncBody {
				node.Body = nil
			}

			if opts.FuncComment {
				node.Decs.Start.Clear()
			}
		}
	}

	dst.Inspect(out, func(node dst.Node) bool {
		if _, exported := Identifier(node); exported && opts.Exported {
			patch(node)
		} else if !exported {
			patch(node)
		}
		return true
	})

	if opts.PackageComment {
		if file, ok := out.(*dst.File); ok {
			file.Decs.Package.Clear()
			file.Decs.Start.Clear()
		}
	}

	return out
}

func Minify[Node dst.Node](node Node, opts MinifyOptions) Node {
	return opts.Minify(node).(Node)
}
