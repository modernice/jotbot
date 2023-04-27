package golang_test

// func TestSkip(t *testing.T) {
// 	cases := []struct {
// 		name  string
// 		files []string
// 		skip  *golang.Skip
// 		want  []string
// 	}{
// 		{
// 			name: "default",
// 			files: []string{
// 				"foo.go",
// 				"foo/foo.go",
// 				"testdata/foo.go",
// 				".foo/foo.go",
// 				".baz/.baz.go",
// 				"bar/baz.go",
// 				"bar/testdata/baz.go",
// 			},
// 			want: []string{
// 				"foo.go",
// 				"foo/foo.go",
// 				"bar/baz.go",
// 			},
// 		},
// 		{
// 			name: "none hidden",
// 			files: []string{
// 				"foo.go",
// 				"bar.go",
// 				"foo/foo.go",
// 				"foo/bar.go",
// 				"testdata/foo.go",
// 				".foo/foo.go",
// 				".baz.go",
// 			},
// 			skip: ptr(golang.SkipNone()),
// 			want: []string{
// 				"foo.go",
// 				"bar.go",
// 				"foo/foo.go",
// 				"foo/bar.go",
// 				"testdata/foo.go",
// 				".foo/foo.go",
// 				".baz.go",
// 			},
// 		},
// 		{
// 			name: "hidden dir",
// 			files: []string{
// 				"foo.go",
// 				".bar/bar.go",
// 				".baz.go", // dotfiles are handled by the Dotfiles option
// 			},
// 			skip: &golang.Skip{Hidden: true},
// 			want: []string{
// 				"foo.go",
// 				".baz.go",
// 			},
// 		},
// 		{
// 			name: "dotfiles",
// 			files: []string{
// 				"foo.go",
// 				".bar.go",
// 			},
// 			skip: &golang.Skip{Dotfiles: true},
// 			want: []string{
// 				"foo.go",
// 			},
// 		},
// 		{
// 			name: "testdata",
// 			files: []string{
// 				"foo.go",
// 				"testdata/foo.go",
// 				"bar/testdata/bar.go",
// 				"bar/bar.go",
// 				".baz.go",
// 			},
// 			skip: &golang.Skip{Testdata: true},
// 			want: []string{
// 				"foo.go",
// 				"bar/bar.go",
// 				".baz.go",
// 			},
// 		},
// 		{
// 			name: "testfiles",
// 			files: []string{
// 				"foo.go",
// 				"foo_test.go",
// 				"bar/bar.go",
// 				"bar/bar_test.go",
// 				".foo/bar/baz.go",
// 				".foo/bar/baz_test.go",
// 			},
// 			skip: &golang.Skip{Testfiles: true},
// 			want: []string{
// 				"foo.go",
// 				"bar/bar.go",
// 				".foo/bar/baz.go",
// 			},
// 		},
// 		{
// 			name: "custom directory skip",
// 			files: []string{
// 				"foo.go",
// 				"foo/foo.go",
// 				"foo/foo/foo.go",
// 				"foo/bar.go",
// 				"foo/bar/bar.go",
// 				"foo/bar/baz.go",
// 				"foo/bar/baz/foo.go",
// 				"foo/what/ever.go",
// 				"bar/baz.go",
// 				"bar/bar/bar.go",
// 				"bar/baz/bar.go",
// 			},
// 			skip: &golang.Skip{
// 				Dir: func(e golang.Exclude) bool {
// 					return e.Path == "foo/bar" || e.DirEntry.Name() == "baz" || e.DirEntry.Name() == "what"
// 				},
// 			},
// 			want: []string{
// 				"foo.go",
// 				"foo/foo.go",
// 				"foo/foo/foo.go",
// 				"foo/bar.go",
// 				"bar/baz.go",
// 				"bar/bar/bar.go",
// 			},
// 		},
// 		{
// 			name: "custom file skip",
// 			files: []string{
// 				"foo.go",
// 				"foo/foo.go",
// 				"bar.go",
// 				"bar/foo.go",
// 				"bar/bar.go",
// 				"bar/baz.go",
// 			},
// 			skip: &golang.Skip{
// 				File: func(e golang.Exclude) bool {
// 					return e.Path == "bar/bar.go" || e.DirEntry.Name() == "foo.go"
// 				},
// 			},
// 			want: []string{
// 				"bar.go",
// 				"bar/baz.go",
// 			},
// 		},
// 	}

// 	for _, tt := range cases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			files := createRepo(t, tt.files)

// 			var opts []golang.FinderOption
// 			if tt.skip != nil {
// 				opts = append(opts, golang.Skip(*tt.skip))
// 			}

// 			findings, err := golang.NewFinder(files, opts...).Uncommented()
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			want := expectedFindings(tt.want)

// 			tests.ExpectFindings(t, want, findings)
// 		})
// 	}
// }

// func createRepo(t *testing.T, files []string) *memfs.FS {
// 	fs := memfs.New()

// 	for _, file := range files {
// 		dir := filepath.Dir(file)
// 		if err := fs.MkdirAll(dir, 0755); err != nil {
// 			t.Fatalf("create dummy dir %s: %v", dir, err)
// 		}
// 		if err := fs.WriteFile(file, []byte("package foo\n\nfunc Foo() {}\n"), 0644); err != nil {
// 			t.Fatalf("create dummy file %s: %v", file, err)
// 		}
// 	}

// 	return fs
// }

// func expectedFindings(files []string) golang.Findings {
// 	out := make(golang.Findings, len(files))
// 	for _, file := range files {
// 		out[file] = []golang.Finding{
// 			{
// 				Path:       file,
// 				Identifier: "Foo",
// 			},
// 		}
// 	}
// 	return out
// }

// func ptr[T any](v T) *T { return &v }
