package find

import (
	"io/fs"
	"strings"
)

// Skip represents a set of rules for excluding directories and files during a
// file system traversal. It contains fields for excluding hidden files,
// dotfiles, test data directories, and test files. It also includes functions
// for customizing directory and file exclusion behavior. Skip is used by Finder
// to avoid traversing unwanted directories and files.
type Skip struct {
	Hidden    bool
	Dotfiles  bool
	Testdata  bool
	Testfiles bool
	Tests     bool

	Dir  func(Exclude) bool
	File func(Exclude) bool
}

// Exclude represents a file or directory that should be excluded from a search.
// It contains a fs.DirEntry and the Path of the file or directory. It is used
// as an argument for Skip's Dir and File functions to determine if a file or
// directory should be skipped during a search.
type Exclude struct {
	fs.DirEntry
	Path string
}

// SkipNone returns a Skip value with all of its fields set to their zero value.
// It is used to indicate that no files or directories should be skipped during
// the search.
func SkipNone() Skip {
	return Skip{}
}

// SkipDefault returns a Skip object with default values set for skipping hidden
// files and directories, dotfiles, testdata, and testfiles. It can be used as a
// starting point for creating custom Skip objects. [Exclude] objects can be
// passed to Skip methods to check if a file or directory should be excluded
// during a file search.
func SkipDefault() Skip {
	return Skip{
		Hidden:    true,
		Dotfiles:  true,
		Testdata:  true,
		Testfiles: true,
	}
}

func (s Skip) apply(f *Finder) {
	f.skip = &s
}

// ExcludeDir returns a boolean indicating whether a directory should be
// excluded from the search. An Exclude struct with the directory's DirEntry and
// Path is passed as an argument. This function checks if the directory name
// starts with a dot when Skip.Hidden is set to true, or if it matches the
// "testdata" folder when Skip.Testdata is set to true. If Skip.Dir is not nil,
// it also calls that function to determine if the directory should be skipped.
func (s Skip) ExcludeDir(e Exclude) bool {
	if s.Hidden && strings.HasPrefix(e.Name(), ".") {
		return true
	}

	if s.Testdata && e.Name() == "testdata" {
		return true
	}

	if s.Dir != nil {
		return s.Dir(e)
	}

	return false
}

// ExcludeFile returns a boolean indicating whether the file should be excluded
// from the search. The exclusion criteria are based on the Skip struct's
// Dotfiles, Testfiles, and File fields. If Dotfiles is true, files with names
// starting with a dot (".") are excluded. If Testfiles is true, files with
// names ending in "_test.go" are excluded. Finally, if the File field of Skip
// is not nil, it will be called with the fs.DirEntry wrapped in an Exclude
// struct to determine whether the file should be excluded.
func (s Skip) ExcludeFile(f Exclude) bool {
	if s.Dotfiles && strings.HasPrefix(f.Name(), ".") {
		return true
	}

	if s.Testfiles && strings.HasSuffix(f.Name(), "_test.go") {
		return true
	}

	if s.File != nil {
		return s.File(f)
	}

	return false
}
