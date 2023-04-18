package tests

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/dave/dst/decorator"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/opendocs/find"
	"github.com/modernice/opendocs/generate"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/modernice/opendocs/internal/slice"
	"github.com/modernice/opendocs/patch"
	"golang.org/x/exp/slices"
)

func ExpectFindings(t *testing.T, want, got find.Findings) {
	t.Helper()

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected findings:\n%s", cmp.Diff(want, got))
	}
}

func ExpectPatch(t *testing.T, want *patch.Patch, got *patch.Patch) {
	t.Helper()

	wantDryRun, err := want.DryRun()
	if err != nil {
		t.Fatalf("dry run 'want': %v", err)
	}

	dryRun, err := got.DryRun()
	if err != nil {
		t.Fatalf("dry run 'got': %v", err)
	}

	if !cmp.Equal(wantDryRun, dryRun) {
		t.Fatalf("dry run mismatch:\n%s", cmp.Diff(wantDryRun, dryRun))
	}
}

func ExpectGenerations(t *testing.T, want, got []generate.Generation) {
	t.Helper()

	less := func(a, b generate.Generation) bool {
		return fmt.Sprintf("%s@%s", a.File, a.Identifier) <= fmt.Sprintf("%s@%s", b.File, b.Identifier)
	}

	slices.SortFunc(want, less)
	slices.SortFunc(got, less)

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected generations:\n%s", cmp.Diff(want, got))
	}
}

func ExpectComment(t *testing.T, identifier, comment string, file io.Reader) {
	root, err := decorator.Parse(file)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	node, ok := nodes.Find(identifier, root)
	if !ok {
		t.Fatalf("find identifier %s: %v", identifier, err)
	}

	comments := slice.Map(node.Decorations().Start.All(), func(c string) string {
		c = strings.TrimPrefix(c, "//")
		return strings.TrimSpace(c)
	})
	foundComment := strings.Join(comments, "\n")

	dif := diff.LineDiff(comment, foundComment)

	if foundComment != comment {
		t.Fatalf("unexpected comment for identifier %s:\n%s\n\nwant:\n%s\n\ngot:\n%s", identifier, dif, comment, foundComment)
	}
}
