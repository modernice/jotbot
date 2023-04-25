package ts_test

import (
	"context"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
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

	tests.ExpectFound(t, []ts.Finding{
		{Identifier: "var:foo", Target: "variable 'foo'"},
		{Identifier: "func:foobar", Target: "function 'foobar'"},
		{Identifier: "iface:Foo", Target: "interface 'Foo'"},
		{Identifier: "prop:Foo.foo", Target: "property 'foo' of 'Foo'"},
		{Identifier: "method:Foo.bar", Target: "method 'bar' of 'Foo'"},
		{Identifier: "prop:Foo.baz", Target: "property 'baz' of 'Foo'"},
		{Identifier: "class:Bar", Target: "class 'Bar'"},
		{Identifier: "prop:Bar.foo", Target: "property 'foo' of 'Bar'"},
		{Identifier: "method:Bar.bar", Target: "method 'bar' of 'Bar'"},
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

	svc := ts.NewFinder(ts.Symbols(ts.Var, ts.Method))

	findings, err := svc.Find(context.Background(), []byte(code))
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []ts.Finding{
		{Identifier: "var:foo", Target: "variable 'foo'"},
		{Identifier: "method:Foo.bar", Target: "method 'bar' of 'Foo'"},
		{Identifier: "method:Bar.bar", Target: "method 'bar' of 'Bar'"},
	}, findings)
}
