package golang_test

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
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

	findings, err := f.Find([]byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectIdentifiers(t, []string{
		"var:X",
		"var:Foo",
		"func:Bar",
		"type:Baz",
		"func:Baz.Baz",
		"type:Foobar",
		"func:Foobar.Foobar",
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

	findings, err := f.Find([]byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectIdentifiers(t, []string{
		"var:Foo",
		"func:Baz",
	}, findings)
}

// X is a struct that represents an empty type used in
// TestFinder_Find_pointerReceiver test function.
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

	findings, err := f.Find([]byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectIdentifiers(t, []string{
		"type:Foo",
		"func:(*Foo).Foo",
		"func:Foo.Bar",
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

	findings, err := f.Find([]byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectIdentifiers(t, []string{
		"func:Foobar",
		"type:Foo",
		"func:Foo.Foo",
		"func:(*Foo).Bar",
		"func:(*Foo).Baz",
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

	findings, err := f.Find([]byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectIdentifiers(t, []string{
		"func:Foobar",
	}, findings)
}

func TestFinder_Find_variableList(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		const (
			Foo = "foo"
			Bar = "bar"
		)
	`)

	f := golang.NewFinder()

	findings, err := f.Find([]byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectIdentifiers(t, []string{"var:Foo", "var:Bar"}, findings)
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

	findings, err := f.Find([]byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectIdentifiers(t, []string{
		"func:TestFoo",
		"func:TestBar",
		"func:Foobar",
	}, findings)
}
