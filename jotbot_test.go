package jotbot_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/internal/tests"
)

func TestJotBot_Find(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "find")
	tests.InitRepo("basic", root)

	bot := jotbot.Default.New(root)

	findings, err := bot.Find(context.Background())
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []jotbot.Finding{
		{File: "foo.go", Finding: find.Finding{Identifier: "Foo", Target: "function 'Foo'"}},
		{File: "bar.go", Finding: find.Finding{Identifier: "Foo", Target: "const 'Foo'"}},
		{File: "bar.go", Finding: find.Finding{Identifier: "Bar", Target: "type 'Bar'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "X", Target: "type 'X'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "X.Foo", Target: "method 'X.Foo'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "(*X).Bar", Target: "method '(*X).Bar'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "Y.Foo", Target: "method 'Y.Foo'"}},
	}, findings)
}
