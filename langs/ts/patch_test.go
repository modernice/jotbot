package ts_test

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/langs/ts"
)

func TestInsertComment(t *testing.T) {
	code := heredoc.Doc(`
		const foo = 'foo'

		export interface Foo {}
	`)

	pos := ts.Position{
		Line:      2,
		Character: 0,
	}

	comment := heredoc.Doc(`
		/**
		 * This is a comment
		 */
	`)

	patched, err := ts.InsertComment(comment, []byte(code), pos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := heredoc.Doc(`
		const foo = 'foo'

		/**
		 * This is a comment
		 */
		export interface Foo {}
	`)

	if string(patched) != want {
		t.Fatalf("unexpected result\n\n%s\n\nwant:\n%s\n\ngot:\n%s", cmp.Diff(want, string(patched)), want, string(patched))
	}
}

func TestInsertComment_indent(t *testing.T) {
	code := heredoc.Doc(`
		export interface Foo {
			foo: string
		}
	`)

	pos := ts.Position{
		Line:      1,
		Character: 1,
	}

	comment := heredoc.Doc(`
		/**
		 * This is a comment
		 */
	`)

	patched, err := ts.InsertComment(comment, []byte(code), pos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := heredoc.Doc(`
		export interface Foo {
			/**
			 * This is a comment
			 */
			foo: string
		}
	`)

	if string(patched) != want {
		t.Fatalf("unexpected result\n\n%s\n\nwant:\n%s\n\ngot:\n%s", cmp.Diff(want, string(patched)), want, string(patched))
	}
}

func TestInsertComment_interfaceMethod(t *testing.T) {
	code := heredoc.Doc(`
		export interface Foo {
			foo(): string
			bar(): number
		}
	`)

	pos := ts.Position{
		Line:      2,
		Character: 1,
	}

	comment := heredoc.Doc(`
		/**
		 * This is a comment
		 */
	`)

	patched, err := ts.InsertComment(comment, []byte(code), pos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := heredoc.Doc(`
		export interface Foo {
			foo(): string
			/**
			 * This is a comment
			 */
			bar(): number
		}
	`)

	if string(patched) != want {
		t.Fatalf("unexpected result\n\n%s\n\nwant:\n%s\n\ngot:\n%s", cmp.Diff(want, string(patched)), want, string(patched))
	}
}
