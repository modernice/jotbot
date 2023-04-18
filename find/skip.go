package find

import (
	"io/fs"
	"strings"
)

type Skip struct {
	Hidden    bool
	Dotfiles  bool
	Testdata  bool
	Testfiles bool
	Tests     bool

	Dir  func(Exclude) bool
	File func(Exclude) bool
}

type Exclude struct {
	fs.DirEntry
	Path string
}

func SkipNone() Skip {
	return Skip{}
}

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
