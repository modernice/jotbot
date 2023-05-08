package golang_test

import (
	"context"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/langs/golang"
	"github.com/modernice/jotbot/patch"
)

var _ interface {
	generate.Language
	patch.Language
	jotbot.Language
} = (*golang.Service)(nil)

func TestService_Patch(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		var Foobar = "foobar"

		func Foo() {}

		type X struct{}

		func (*X) Foo() {}

		type Y interface {
			Foo() string
		}
	`)

	svc := golang.Must()

	patched, err := svc.Patch(context.Background(), "var:Foobar", "Foobar is foobar.", []byte(code))
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	patched, err = svc.Patch(context.Background(), "func:Foo", "Foo is a foo.", patched)
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	patched, err = svc.Patch(context.Background(), "type:X", "X is a x.", patched)
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	patched, err = svc.Patch(context.Background(), "func:(*X).Foo", "Foo is a foo.", patched)
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	patched, err = svc.Patch(context.Background(), "type:Y", "Y is a y.", patched)
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	patched, err = svc.Patch(context.Background(), "func:Y.Foo", "Foo is a foo.", patched)
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	expect := heredoc.Doc(`
		package foo

		// Foobar is foobar.
		var Foobar = "foobar"

		// Foo is a foo.
		func Foo() {}

		// X is a x.
		type X struct{}

		// Foo is a foo.
		func (*X) Foo() {}

		// Y is a y.
		type Y interface {
			// Foo is a foo.
			Foo() string
		}
	`)

	if string(patched) != expect {
		t.Errorf("Patch() returned invalid code:\n\n%s\n\n%s", cmp.Diff(expect, string(patched)), string(patched))
	}
}

func TestService_Patch_groupDeclaration(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		const (
			Foo = "foo"
			Bar = "bar"
		)
	`)

	svc := golang.Must()

	patched, err := svc.Patch(context.Background(), "var:Foo", "Foo is a foo.", []byte(code))
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	patched, err = svc.Patch(context.Background(), "var:Bar", "Bar is a bar.", patched)
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	expect := heredoc.Doc(`
		package foo

		const (
			// Foo is a foo.
			Foo = "foo"

			// Bar is a bar.
			Bar = "bar"
		)
	`)

	if string(patched) != expect {
		t.Errorf("Patch() returned invalid code:\n\n%s\n\n%s", cmp.Diff(expect, string(patched)), string(patched))
	}
}

func TestService_Patch_interfaceMethods(t *testing.T) {
	code := heredoc.Doc(`
		package foo

		type Foo interface {
			Bar()
			Baz()
		}
	`)

	svc := golang.Must()

	patched, err := svc.Patch(context.Background(), "func:Foo.Bar", "Bar is a bar.", []byte(code))
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}
	patched, err = svc.Patch(context.Background(), "func:Foo.Baz", "Baz is a baz.", patched)
	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	expect := heredoc.Doc(`
		package foo

		type Foo interface {
			// Bar is a bar.
			Bar()

			// Baz is a baz.
			Baz()
		}
	`)

	if string(patched) != expect {
		t.Errorf("Patch() returned invalid code:\n\n%s\n\n%s", cmp.Diff(expect, string(patched)), string(patched))
	}
}
