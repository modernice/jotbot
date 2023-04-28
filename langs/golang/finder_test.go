package golang_test

import (
	"context"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/internal/tests"
	"github.com/modernice/jotbot/langs/golang"
)

func TestFinder_Find(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		const X = "x"

		var Foo = "foo"

		func Bar() string {
			return "bar"
		}

		type Baz struct{}

		func (Baz) Baz() string {
			return "baz"
		}

		type Foobar interface {
			Foobar() string
		}
	`)

	f := golang.NewFinder()

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "X", Target: "const 'X'"},
		{Identifier: "Foo", Target: "variable 'Foo'"},
		{Identifier: "Bar", Target: "function 'Bar'"},
		{Identifier: "Baz", Target: "type 'Baz'"},
		{Identifier: "Baz.Baz", Target: "method 'Baz.Baz'"},
		{Identifier: "Foobar", Target: "type 'Foobar'"},
		{Identifier: "Foobar.Foobar", Target: "method 'Foobar.Foobar'"},
	}, findings)
}

func TestFinder_Find_onlyUncommented(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		const Foo = "foo"

		// Bar is a variable.
		var Bar = "bar"

		func Baz() string {
			return Foo
		}

		// Foobar is a struct.
		type Foobar struct{}
	`)

	f := golang.NewFinder()

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "Foo", Target: "const 'Foo'"},
		{Identifier: "Baz", Target: "function 'Baz'"},
	}, findings)
}

type X struct{}

func TestFinder_Find_pointerReceiver(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		type Foo struct{}

		func (f *Foo) Foo() string {
			return "foo"
		}

		func (f Foo) Bar() string {
			return "bar"
		}
	`)

	f := golang.NewFinder()

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "Foo", Target: "type 'Foo'"},
		{Identifier: "(*Foo).Foo", Target: "method '(*Foo).Foo'"},
		{Identifier: "Foo.Bar", Target: "method 'Foo.Bar'"},
	}, findings)
}

func TestFinder_Find_generics(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		func Foobar[T any, T2 string](t T, t2 T2) (T, T2) {
			return t, t2
		}

		type Foo[T any] struct {
			Bar T
		}

		func (f Foo[T]) Foo() T {
			return f.Bar
		}

		func (f *Foo[T]) Bar() T {
			return f.Bar
		}

		func (*Foo[_]) Baz() string {
			return "baz"
		}
	`)

	f := golang.NewFinder()

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "Foobar", Target: "function 'Foobar'"},
		{Identifier: "Foo", Target: "type 'Foo'"},
		{Identifier: "Foo.Foo", Target: "method 'Foo.Foo'"},
		{Identifier: "(*Foo).Bar", Target: "method '(*Foo).Bar'"},
		{Identifier: "(*Foo).Baz", Target: "method '(*Foo).Baz'"},
	}, findings)
}

func TestFinder_Find_excludesTestsByDefault(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		import "testing"

		func TestFoo() {}

		func TestBar(t *testing.T) {}

		func Foobar() {}
	`)

	f := golang.NewFinder()

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "Foobar", Target: "function 'Foobar'"},
	}, findings)
}

func TestFindTests(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		import "testing"

		func TestFoo() {}

		func TestBar(t *testing.T) {}

		func Foobar() {}
	`)

	f := golang.NewFinder(golang.FindTests(true))

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "TestFoo", Target: "function 'TestFoo'"},
		{Identifier: "TestBar", Target: "function 'TestBar'"},
		{Identifier: "Foobar", Target: "function 'Foobar'"},
	}, findings)
}
