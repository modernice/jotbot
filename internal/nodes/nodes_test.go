package nodes_test

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/dave/dst"
	"github.com/modernice/jotbot/internal/nodes"
)

func TestFind(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		func Foo() {}

		type Bar struct{}

		func (*Bar) Bar() {}
	`)

	root := nodes.MustParse(code)

	{
		_, node, ok := nodes.Find("func:Foo", root)
		if !ok {
			t.Fatalf("Find() failed to find func:Foo")
		}

		if node.(*dst.FuncDecl).Name.Name != "Foo" {
			t.Fatalf("Find() returned wrong declaration; want %q, got %q", "Foo", node.(*dst.FuncDecl).Name.Name)
		}
	}

	{
		spec, node, ok := nodes.Find("type:Bar", root)
		if !ok {
			t.Fatalf("Find() failed to find type:Bar")
		}

		if spec.(*dst.TypeSpec).Name.Name != "Bar" {
			t.Fatalf("Find() returned wrong spec; want %q, got %q", "Bar", spec.(*dst.TypeSpec).Name.Name)
		}

		if node.(*dst.GenDecl).Specs[0] != spec {
			t.Fatalf("Find() returned wrong declaration; want %v, got %v", spec, node.(*dst.GenDecl).Specs[0])
		}
	}

	{
		_, node, ok := nodes.Find("func:(*Bar).Bar", root)
		if !ok {
			t.Fatalf("Find() failed to find func:(*Bar).Bar")
		}

		if node.(*dst.FuncDecl).Name.Name != "Bar" {
			t.Fatalf("Find() returned wrong declaration; want %q, got %q", "Bar", node.(*dst.FuncDecl).Name.Name)
		}
	}
}

func TestCommentTarget(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		const (
			Foo = "foo"
			Bar = "bar"
		)

		var (
			foobar = "foobar"
			barbaz = "barbaz"
		)

		type (
			X struct{}
			Y interface{}
		)

		func FooFn() {}

		func (*X) BarFn() {}
	`)

	root := nodes.MustParse(code)

	{
		spec, decl, ok := nodes.FindValue("var:Foo", root)
		if !ok {
			t.Fatalf("FindValue() failed to find var:Foo")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != spec {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", spec, target)
		}
	}

	{
		spec, decl, ok := nodes.FindValue("var:Foo", root)
		if !ok {
			t.Fatalf("FindValue() failed to find var:Foo")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != spec {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", spec, target)
		}
	}

	{
		spec, decl, ok := nodes.FindValue("var:foobar", root)
		if !ok {
			t.Fatalf("FindValue() failed to find var:foobar")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != spec {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", spec, target)
		}
	}

	{
		spec, decl, ok := nodes.FindValue("var:barbaz", root)
		if !ok {
			t.Fatalf("FindValue() failed to find var:barbaz")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != spec {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", spec, target)
		}
	}

	{
		spec, decl, ok := nodes.FindType("type:X", root)
		if !ok {
			t.Fatalf("FindType() failed to find type:X")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != spec {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", spec, target)
		}
	}

	{
		spec, decl, ok := nodes.FindType("type:Y", root)
		if !ok {
			t.Fatalf("FindType() failed to find type:Y")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != spec {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", spec, target)
		}
	}

	{
		spec, decl, ok := nodes.Find("func:FooFn", root)
		if !ok {
			t.Fatalf("Find() failed to find func:FooFn")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != decl {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", decl, target)
		}
	}

	{
		spec, decl, ok := nodes.Find("func:(*X).BarFn", root)
		if !ok {
			t.Fatalf("Find() failed to find func:(*X).BarFn")
		}

		target := nodes.CommentTarget(spec, decl)
		if target != decl {
			t.Fatalf("CommentTarget() returned wrong target; want %v, got %v", decl, target)
		}
	}
}
