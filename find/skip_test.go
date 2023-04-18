package find_test

import (
	"path/filepath"
	"testing"

	"github.com/modernice/opendocs/find"
	"github.com/modernice/opendocs/internal/tests"
	"github.com/psanford/memfs"
)

func TestSkip(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		skip  *find.Skip
		want  []string
	}{
		{
			name: "default",
			files: []string{
				"foo.go",
				"foo/foo.go",
				"testdata/foo.go",
				".foo/foo.go",
				".baz/.baz.go",
				"bar/baz.go",
				"bar/testdata/baz.go",
			},
			want: []string{
				"foo.go",
				"foo/foo.go",
				"bar/baz.go",
			},
		},
		{
			name: "none hidden",
			files: []string{
				"foo.go",
				"bar.go",
				"foo/foo.go",
				"foo/bar.go",
				"testdata/foo.go",
				".foo/foo.go",
				".baz.go",
			},
			skip: ptr(find.SkipNone()),
			want: []string{
				"foo.go",
				"bar.go",
				"foo/foo.go",
				"foo/bar.go",
				"testdata/foo.go",
				".foo/foo.go",
				".baz.go",
			},
		},
		{
			name: "hidden dir",
			files: []string{
				"foo.go",
				".bar/bar.go",
				".baz.go", // dotfiles are handled by the Dotfiles option
			},
			skip: &find.Skip{Hidden: true},
			want: []string{
				"foo.go",
				".baz.go",
			},
		},
		{
			name: "dotfiles",
			files: []string{
				"foo.go",
				".bar.go",
			},
			skip: &find.Skip{Dotfiles: true},
			want: []string{
				"foo.go",
			},
		},
		{
			name: "testdata",
			files: []string{
				"foo.go",
				"testdata/foo.go",
				"bar/testdata/bar.go",
				"bar/bar.go",
				".baz.go",
			},
			skip: &find.Skip{Testdata: true},
			want: []string{
				"foo.go",
				"bar/bar.go",
				".baz.go",
			},
		},
		{
			name: "testfiles",
			files: []string{
				"foo.go",
				"foo_test.go",
				"bar/bar.go",
				"bar/bar_test.go",
				".foo/bar/baz.go",
				".foo/bar/baz_test.go",
			},
			skip: &find.Skip{Testfiles: true},
			want: []string{
				"foo.go",
				"bar/bar.go",
				".foo/bar/baz.go",
			},
		},
		{
			name: "custom directory skip",
			files: []string{
				"foo.go",
				"foo/foo.go",
				"foo/foo/foo.go",
				"foo/bar.go",
				"foo/bar/bar.go",
				"foo/bar/baz.go",
				"foo/bar/baz/foo.go",
				"foo/what/ever.go",
				"bar/baz.go",
				"bar/bar/bar.go",
				"bar/baz/bar.go",
			},
			skip: &find.Skip{
				Dir: func(e find.Exclude) bool {
					return e.Path == "foo/bar" || e.DirEntry.Name() == "baz" || e.DirEntry.Name() == "what"
				},
			},
			want: []string{
				"foo.go",
				"foo/foo.go",
				"foo/foo/foo.go",
				"foo/bar.go",
				"bar/baz.go",
				"bar/bar/bar.go",
			},
		},
		{
			name: "custom file skip",
			files: []string{
				"foo.go",
				"foo/foo.go",
				"bar.go",
				"bar/foo.go",
				"bar/bar.go",
				"bar/baz.go",
			},
			skip: &find.Skip{
				File: func(e find.Exclude) bool {
					return e.Path == "bar/bar.go" || e.DirEntry.Name() == "foo.go"
				},
			},
			want: []string{
				"bar.go",
				"bar/baz.go",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			files := createRepo(t, tt.files)

			var opts []find.Option
			if tt.skip != nil {
				opts = append(opts, find.Skip(*tt.skip))
			}

			findings, err := find.New(files, opts...).Uncommented()
			if err != nil {
				t.Fatal(err)
			}

			want := expectedFindings(tt.want)

			tests.ExpectFindings(t, want, findings)
		})
	}
}

func createRepo(t *testing.T, files []string) *memfs.FS {
	fs := memfs.New()

	for _, file := range files {
		dir := filepath.Dir(file)
		if err := fs.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("create dummy dir %s: %v", dir, err)
		}
		if err := fs.WriteFile(file, []byte("package foo\n\nfunc Foo() {}\n"), 0644); err != nil {
			t.Fatalf("create dummy file %s: %v", file, err)
		}
	}

	return fs
}

func expectedFindings(files []string) find.Findings {
	out := make(find.Findings, len(files))
	for _, file := range files {
		out[file] = []find.Finding{
			{
				Path:       file,
				Identifier: "Foo",
			},
		}
	}
	return out
}

func ptr[T any](v T) *T { return &v }
