package ts_test

import (
	"context"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/internal/tests"
	"github.com/modernice/jotbot/langs/ts"
)

func TestFinder_Find(t *testing.T) {
	code := heredoc.Doc(`
		export const foo = 'foo'

		export function foobar() {
			return 'foobar'
		}

		export interface Foo {
			foo: string
			bar(): void
			baz: () => number
		}

		export class Bar {
			foo = 'foo'
			bar() {}
			private baz() {}
			#foobar() {}
		}
	`)

	f := ts.NewFinder()

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "var:foo"},
		{Identifier: "func:foobar"},
		{Identifier: "iface:Foo"},
		{Identifier: "prop:Foo.foo"},
		{Identifier: "method:Foo.bar"},
		{Identifier: "prop:Foo.baz"},
		{Identifier: "class:Bar"},
		{Identifier: "prop:Bar.foo"},
		{Identifier: "method:Bar.bar"},
	}, findings)
}

func TestSymbols(t *testing.T) {
	code := heredoc.Doc(`
		export const foo = 'foo'

		export function foobar() {
			return 'foobar'
		}

		export interface Foo {
			foo: string
			bar(): void
			baz: () => number
		}

		export class Bar {
			foo = 'foo'
			bar() {}
			private baz() {}
			#foobar() {}
		}
	`)

	f := ts.NewFinder(ts.Symbols(ts.Var, ts.Method))

	findings, err := f.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []find.Finding{
		{Identifier: "var:foo"},
		{Identifier: "method:Foo.bar"},
		{Identifier: "method:Bar.bar"},
	}, findings)
}

func TestFinder_Position(t *testing.T) {
	code := heredoc.Doc(`
		export const foo = 'foo'

		export function bar() {}
	`)

	f := ts.NewFinder()

	pos, err := f.Position(context.Background(), "func:bar", []byte(code))
	if err != nil {
		t.Fatalf("Position() failed: %v", err)
	}

	if pos.Line != 2 {
		t.Errorf("Position() returned wrong line; want %d; got %d", 2, pos.Line)
	}

	if pos.Character != 0 {
		t.Errorf("Position() returned wrong character; want %d; got %d", 9, pos.Character)
	}
}
