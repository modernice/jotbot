package find

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/modernice/opendocs/internal/slice"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

// Finder represents a code finder that searches for Go types and functions in a
// file system. It provides methods to search for all types/functions or only
// those without comments. To create a new Finder, use the New function with an
// fs.FS and optional Options.
type Finder struct {
	repo  fs.FS
	skip  *Skip
	globs []string
	log   *slog.Logger
}

// Finding represents a code finder that searches for types and functions. It
// has methods to find all types/functions or only those without comments. It
// can also exclude files and directories based on glob patterns.
type Finding struct {
	Path       string
	Identifier string
}

// String returns a string representation of the Finding in the format
// "Path@Identifier".
func (f Finding) String() string {
	return fmt.Sprintf("%s@%s", f.Path, f.Identifier)
}

// Findings represents a map of string slices containing the path and identifier
// of types and functions found in a repository.
type Findings map[string][]Finding

// Option is an interface that allows for optional configuration of a Finder
// instance. An Option is implemented as a function that modifies a *Finder
// instance.
type Option interface {
	apply(*Finder)
}

type optionFunc func(*Finder)

func (opt optionFunc) apply(f *Finder) {
	opt(f)
}

// WithLogger returns an Option that sets the logger for a Finder. The logger is
// used to log information about the search performed by the Finder.
func WithLogger(h slog.Handler) Option {
	return optionFunc(func(f *Finder) {
		f.log = slog.New(h)
	})
}

// Glob adds a file glob pattern to the Finder options. The Finder will only
// search files that match the glob pattern(s) specified.
func Glob(pattern ...string) Option {
	pattern = slice.Map(pattern, strings.TrimSpace)
	pattern = slice.NoZero(pattern)
	return optionFunc(func(f *Finder) {
		f.globs = append(f.globs, pattern...)
	})
}

// New creates a new Finder with the given fs.FS and Option(s). It returns the
// new Finder.
func New(repo fs.FS, opts ...Option) *Finder {
	f := &Finder{repo: repo}
	for _, opt := range opts {
		opt.apply(f)
	}
	if f.skip == nil {
		skip := SkipDefault()
		f.skip = &skip
	}
	if f.log == nil {
		f.log = internal.NopLogger()
	}
	return f
}

// All returns a map of Findings for all types and functions in the repository
// that match the Finder's globs.
func (f *Finder) All() (Findings, error) {
	return f.find(true)
}

// Uncommented returns the uncommented types and functions found in the source
// files.
func (f *Finder) Uncommented() (Findings, error) {
	return f.find(false)
}

func (f *Finder) find(all bool) (Findings, error) {
	if all {
		f.log.Info("Searching for code in repository ...", "repo", f.repo)
	} else {
		f.log.Info("Searching for uncommented code in repository ...", "repo", f.repo)
	}

	allFindings := make(Findings)

	globExclude, err := f.parseGlobOptions()
	if err != nil {
		return nil, fmt.Errorf("parse glob options: %w", err)
	}

	if err := fs.WalkDir(f.repo, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "." || d.Name() == "" {
			return nil
		}

		exclude := Exclude{
			DirEntry: d,
			Path:     path,
		}

		if d.IsDir() {
			if f.skip != nil && f.skip.ExcludeDir(exclude) {
				f.log.Debug("Skipping directory", "dir", path)
				return filepath.SkipDir
			}
			return nil
		}

		if !isGoFile(d) {
			f.log.Debug("Skipping file", "path", path, "reason", "not a Go file")
			return nil
		}

		if globExclude(path) {
			f.log.Debug("Skipping file", "path", path, "reason", "glob")
			return nil
		}

		if f.skip != nil && f.skip.ExcludeFile(exclude) {
			f.log.Debug("Skipping file", "path", path)
			return nil
		}

		var findings []Finding
		if all {
			findings, err = f.findInFile(path, true)
		} else {
			findings, err = f.findInFile(path, false)
		}
		if err != nil {
			return fmt.Errorf("find code in %s: %w", path, err)
		}

		allFindings[path] = append(allFindings[path], findings...)

		return nil
	}); err != nil {
		return nil, err
	}

	return allFindings, nil
}

func (f *Finder) parseGlobOptions() (func(string) bool, error) {
	if len(f.globs) == 0 {
		return func(string) bool { return false }, nil
	}

	var allowed []string
	for _, pattern := range f.globs {
		files, err := doublestar.Glob(f.repo, pattern)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", pattern, err)
		}
		allowed = append(allowed, files...)
	}
	allowed = slice.Unique(allowed)

	return func(path string) bool {
		return !slices.Contains(allowed, filepath.Clean(path))
	}, nil
}

func (f *Finder) findInFile(path string, all bool) ([]Finding, error) {
	f.log.Info(fmt.Sprintf("Searching for uncommented code in %s ...", path))

	var findings []Finding

	codeFile, err := f.repo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	defer codeFile.Close()

	code, err := io.ReadAll(codeFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	fset := token.NewFileSet()
	node, err := decorator.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}

	for _, node := range node.Decls {
		var (
			identifier string
			exported   bool
		)

		switch node := node.(type) {
		case *dst.FuncDecl:
			if !all && nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			identifier, exported = nodes.Identifier(node)
		case *dst.GenDecl:
			if !all && nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			if len(node.Specs) == 0 {
				break
			}

			spec := node.Specs[0]

			switch spec := spec.(type) {
			case *dst.TypeSpec:
				if all || !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
					identifier, exported = nodes.Identifier(spec)
				}
			case *dst.ValueSpec:
				if all || !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
					identifier, exported = nodes.Identifier(spec)
				}
			}
		}

		if exported && identifier != "" && identifier != "_" {
			findings = append(findings, Finding{
				Path:       path,
				Identifier: identifier,
			})
		}
	}

	slices.SortFunc(findings, func(a, b Finding) bool {
		if a.Path < b.Path {
			return true
		}

		if a.Path == b.Path {
			return a.Identifier <= b.Identifier
		}

		return false
	})

	idents := slice.Map(findings, func(f Finding) string { return f.Identifier })

	f.log.Info(fmt.Sprintf("Found %d uncommented types/functions in %s", len(findings), path), "identifiers", idents)

	return findings, nil
}

func isGoFile(d fs.DirEntry) bool {
	return filepath.Ext(d.Name()) == ".go"
}
