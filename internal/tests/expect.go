package tests

import (
	"io"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/dave/dst/decorator"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal/nodes"
	"github.com/modernice/jotbot/internal/slice"
	"github.com/modernice/jotbot/langs/golang"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func ExpectFiles(t *testing.T, want, got []string) {
	t.Helper()

	slices.Sort(want)
	slices.Sort(got)

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected files:\n%s", cmp.Diff(want, got))
	}
}

func ExpectFound[Finding interface{ String() string }](t *testing.T, want, got []Finding) {
	t.Helper()

	less := func(a, b Finding) bool {
		return a.String() <= b.String()
	}

	slices.SortFunc(want, less)
	slices.SortFunc(got, less)

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected findings:\n%s\n\nwant:\n%v\n\ngot:\n%v", cmp.Diff(want, got), want, got)
	}
}

// ExpectFindings tests if two find.Findings are equal. It sorts the findings by
// identifier, then compares the sorted findings using go-cmp/cmp.
func ExpectFindings[Findings ~map[string][]Finding, Finding interface{ GetIdentifier() string }](t *testing.T, want, got Findings) {
	t.Helper()

	want = maps.Clone(want)
	got = maps.Clone(got)

	sort := func(findings []Finding) {
		slices.SortFunc(findings, func(a, b Finding) bool {
			return a.GetIdentifier() <= b.GetIdentifier()
		})
	}

	for file, findings := range want {
		sort(findings)
		want[file] = findings
	}
	for file, findings := range got {
		sort(findings)
		got[file] = findings
	}

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected findings:\n%s", cmp.Diff(want, got))
	}
}

// func ExpectPatch(t *testing.T, want, got generate.Patch) {
// 	t.Helper()

// 	want = maps.Clone(want)
// 	got = maps.Clone(got)

// 	sortPatch(want)
// 	sortPatch(got)

// 	if !cmp.Equal(want, got) {
// 		t.Fatalf("unexpected patch:\n%s", cmp.Diff(want, got))
// 	}
// }

// ExpectGoPatch compares two *golang.Patch structs and fails the test if their
// DryRun output differs.
func ExpectGoPatch(t *testing.T, want *golang.Patch, got *golang.Patch) {
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

func ExpectGeneratedFiles(t *testing.T, want, got []generate.File) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("unexpected number of generated files:\n%s", cmp.Diff(want, got))
	}

	less := func(a, b generate.File) bool { return a.Path <= b.Path }
	slices.SortFunc(want, less)

	for i, file := range got {
		want := want[i]
		if want.Path != file.Path {
			t.Fatalf("unexpected generated file:\n%s", cmp.Diff(want, file))
		}
		ExpectGenerations(t, want.Docs, file.Docs)
	}
}

// ExpectGenerations compares two slices of generate.Generation and fails the
// test if they are not equal.
func ExpectGenerations(t *testing.T, want, got []generate.Documentation) {
	t.Helper()

	less := func(a, b generate.Documentation) bool { return a.Identifier <= b.Identifier }

	slices.SortFunc(want, less)
	slices.SortFunc(got, less)

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected generations:\n%s", cmp.Diff(want, got))
	}
}

// ExpectComment verifies that the comment for the specified identifier in the
// provided file matches the expected comment. If the comments do not match,
// ExpectComment will fail the test with a detailed diff of the expected and
// found comments.
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

// func sortPatch(patch generate.Patch) {
// 	for file, edits := range patch {
// 		slices.SortFunc(edits, func(a, b generate.Generation) bool {
// 			return a.Identifier <= b.Identifier
// 		})
// 		patch[file] = edits
// 	}
// }
