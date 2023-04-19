package find

import (
	"io/fs"
	"strings"
)

// Skip represents a set of criteria for excluding directories and files from a
// search. It contains fields for excluding hidden files and directories,
// dotfiles, testdata, test files, and directories matching a custom function.
// The Dir and File fields can be set to custom functions for excluding
// directories and files respectively. The ExcludeDir and ExcludeFile methods
// use the Skip criteria to determine whether to exclude a directory or file.
type Skip struct {
	Hidden    bool
	Dotfiles  bool
	Testdata  bool
	Testfiles bool
	Tests     bool

	Dir  func(Exclude) bool
	File func(Exclude) bool
}

// Exclude represents a set of exclusion rules for file and directory names. It
// contains methods to exclude directories and files based on certain criteria
// such as hidden files, test data, and custom rules. The Exclude struct is used
// by the Skip struct to apply exclusion rules to a Finder.
type Exclude struct {
	fs.DirEntry
	Path string
}

// SkipNone returns a Skip struct with all fields set to their zero values. It
// can be used as a default Skip value if no exclusions are needed.
func SkipNone() Skip {
	return Skip{}
}

// SkipDefault returns a Skip value with the default values set for skipping
// hidden files and directories, dotfiles, testdata directories, and test files.
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

// ExcludeDir returns a boolean value indicating whether a directory should be
// excluded from the search. It takes an Exclude struct as input, which contains
// information about the directory being evaluated.
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

// ExcludeFile returns a boolean value indicating whether the given file should
// be excluded from the search based on Skip's Dotfiles, Testfiles, and File
// fields.
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
