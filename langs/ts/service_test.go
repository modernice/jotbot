package ts_test

import (
	"context"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/langs/ts"
)

var _ jotbot.Language = (*ts.Service)(nil)

func TestService_Patch_interfaceFields(t *testing.T) {
	code := heredoc.Doc(`
		export interface Foo {
			foo: string
			bar(): string
		}
	`)

	comment := "This is a comment"

	svc := ts.New()

	patched, err := svc.Patch(context.Background(), "prop:Foo.foo", comment, []byte(code))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := heredoc.Doc(`
		export interface Foo {
			/** This is a comment */
			foo: string
			bar(): string
		}
	`)

	if string(patched) != want {
		t.Fatalf("unexpected result\n\n%s\n\nwant:\n%s\n\ngot:\n%s", cmp.Diff(want, string(patched)), want, string(patched))
	}

	patched, err = svc.Patch(context.Background(), "method:Foo.bar", comment, patched)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want = heredoc.Doc(`
		export interface Foo {
			/** This is a comment */
			foo: string
			/** This is a comment */
			bar(): string
		}
	`)

	if string(patched) != want {
		t.Fatalf("unexpected result\n\n%s\n\nwant:\n%s\n\ngot:\n%s", cmp.Diff(want, string(patched)), want, string(patched))
	}
}

func TestNormalizeGeneratedComment(t *testing.T) {
	comment := heredoc.Doc(`
		/**
		 * This is a multi-line
		 * comment. I am very short in width.
		 *
		 * But I have multiple
		 * paragraphs.
		 *
		 * @param foo This is a parameter.
		 * @return This is a return value.
		 */
	`)

	normalized := ts.NormalizeGeneratedComment(comment)

	want := "This is a multi-line comment. I am very short in width.\n\nBut I have multiple paragraphs."

	if normalized != want {
		t.Fatalf("unexpected result\n\n%s\n\nwant:\n%s\n\ngot:\n%s", cmp.Diff(want, normalized), want, normalized)
	}
}
