package nodes

import (
	"github.com/dave/dst"
)

const (
	MinifyFuncBody MinifyFlags = 1 << iota
	MinifyFuncComment
)

type MinifyFlags uint

type MinifyOptions struct {
	Exported, Unexported MinifyFlags
	PackageComment       bool
}

func (flags MinifyFlags) FuncBody() bool {
	return flags&MinifyFuncBody != 0
}

func (flags MinifyFlags) Comment() bool {
	return flags&MinifyFuncComment != 0
}

func (opts MinifyOptions) Minify(node dst.Node) dst.Node {
	out := dst.Clone(node)

	patch := func(node dst.Node, flags MinifyFlags) {
		switch node := node.(type) {
		case *dst.FuncDecl:
			if flags.FuncBody() {
				node.Body = nil
			}

			if flags.Comment() {
				node.Decs.Start.Clear()
			}
		}
	}

	dst.Inspect(out, func(node dst.Node) bool {
		if _, exported := Identifier(node); exported {
			patch(node, opts.Exported)
		} else {
			patch(node, opts.Unexported)
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

func MinifyUnexported[Node dst.Node](node Node, flags ...MinifyFlags) Node {
	f := MinifyFuncBody | MinifyFuncComment
	for _, flag := range flags {
		f |= flag
	}
	return Minify(node, MinifyOptions{Unexported: f})
}

func MinifyExported[Node dst.Node](node Node, flags ...MinifyFlags) Node {
	f := MinifyFuncBody | MinifyFuncComment
	for _, flag := range flags {
		f |= flag
	}
	return Minify(node, MinifyOptions{
		Exported:   f,
		Unexported: f,
	})
}

func MinifyAll[Node dst.Node](node Node, flags ...MinifyFlags) Node {
	f := MinifyFuncBody | MinifyFuncComment
	for _, flag := range flags {
		f |= flag
	}
	return Minify(node, MinifyOptions{
		Exported:       f,
		Unexported:     f,
		PackageComment: true,
	})
}
