package find

import (
	"io/fs"
	"strings"
)

// Skip represents a set of rules for excluding directories and files during a
// file search. It contains fields for excluding hidden files, dotfiles,
// testdata directories, test files, and directories and files specified by
// custom functions. The SkipNone function returns an empty Skip struct, while
// the SkipDefault function returns a Skip struct with default exclusion rules.
// The apply method applies the Skip rules to a Finder struct. The ExcludeDir
// method returns true if the directory should be excluded based on the Skip
// rules, and the ExcludeFile method returns true if the file should be excluded
// based on the Skip rules.
type Skip struct {
	Hidden    bool
	Dotfiles  bool
	Testdata  bool
	Testfiles bool
	Tests     bool

	Dir  func(Exclude) bool
	File func(Exclude) bool
}

// Exclude represents a struct that can be used to exclude directories and files
// from a search. It contains methods to exclude directories and files based on
// various criteria such as hidden files, test data, and custom functions. It is
// used in conjunction with the Skip struct, which contains options for
// excluding directories and files.
type Exclude struct {
	fs.DirEntry
	Path string
}

// SkipNone returns a Skip struct with all fields set to their zero values. It
// is used to indicate that no files or directories should be skipped during a
// search.
func SkipNone() Skip {
	return Skip{}
}

// SkipDefault returns a Skip struct with default values for excluding files and
// directories. It excludes hidden files and directories, dotfiles, testdata
// directories, and test files.
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
// excluded from the search. It takes an Exclude struct as input, which contains
// information about the directory being evaluated. The Skip struct that calls
// this method contains several boolean fields that determine whether certain
// types of directories should be excluded, such as hidden directories and
// testdata directories. If the Skip struct also contains a Dir function, it
// will be called to determine whether the directory should be excluded.
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

// ExcludeFile returns a boolean indicating whether the given file should be
// excluded based on the Skip configuration. If Dotfiles is true and the file
// name starts with a ".", it will be excluded. If Testfiles is true and the
// file name ends with "_test.go", it will be excluded. If File is not nil, it
// will be called with the given Exclude and its result will be returned.
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
