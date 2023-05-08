package find

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	// DefaultExtensions is a predefined slice of strings containing common file
	// extensions (such as ".go" and ".ts") used as default values when searching
	// for files using the Options struct.
	DefaultExtensions = []string{
		".go",
		".ts",
	}

	// DefaultExclude is a predefined slice of strings containing common file
	// exclusion patterns, such as hidden directories, build directories, and test
	// directories. It is used as a default value when searching for files using the
	// Options struct.
	DefaultExclude = []string{
		"**/.*/**",           // hidden directories
		"**/dist/**",         // node builds
		"**/node_modules/**", // node dependencies
		"**/vendor/**",       // vendored dependencies
		"**/testdata/**",     // tests
		"**/test/**",         // tests
		"**/tests/**",        // tests
	}

	// Default is a predefined instance of the Options struct with commonly used
	// file extensions and exclusion patterns. It serves as a starting point for
	// file searching configurations.
	Default = Options{
		Extensions: DefaultExtensions,
		Exclude:    DefaultExclude,
	}
)

// Options is a configuration struct that defines the filtering rules for file
// searching, such as file extensions to include, patterns to include or
// exclude. It provides methods for determining if a given file path is included
// or excluded based on these rules.
type Options struct {
	Extensions []string
	Include    []string
	Exclude    []string
}

// Option is a functional option type that allows customization of the behavior
// of the [Options] struct, which is used in the file search process. It can be
// used to modify extensions, include or exclude specific patterns, and other
// search-related configurations.
type Option func(*Options)

// Extensions returns an Option that sets the allowed file extensions for the
// Options struct. The given exts parameter is a list of strings representing
// the desired file extensions.
func Extensions(exts ...string) Option {
	return func(o *Options) {
		o.Extensions = exts
	}
}

// Include adds the given patterns to the list of include patterns for file
// search, allowing the inclusion of matching files in the search results.
func Include(patterns ...string) Option {
	return func(o *Options) {
		o.Include = append(o.Include, patterns...)
	}
}

// Exclude appends the given patterns to the Options.Exclude field, marking them
// to be excluded from the file search. The function returns an Option for use
// with the Files function.
func Exclude(patterns ...string) Option {
	return func(o *Options) {
		o.Exclude = append(o.Exclude, patterns...)
	}
}

// Files returns a list of file paths from the provided fs.FS, filtered based on
// the given options. The options can include or exclude files based on file
// extensions, and include or exclude patterns. The function also supports
// context cancellation.
func Files(ctx context.Context, files fs.FS, opts ...Option) ([]string, error) {
	cfg := Default
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg.Find(ctx, files)
}

// Find searches the provided file system (fs.FS) using the options specified,
// such as extensions, include and exclude patterns, and returns a slice of file
// paths matching the criteria. It also respects the context (context.Context)
// for cancellation or timeouts.
func (f Options) Find(ctx context.Context, files fs.FS) ([]string, error) {
	if len(f.Extensions) == 0 {
		f.Extensions = DefaultExtensions
	}

	var found []string
	if err := fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if f.excluded(path) {
				return fs.SkipDir
			}
			return nil
		}

		if !f.included(path) {
			return nil
		}

		if f.excluded(path) {
			return nil
		}

		found = append(found, path)

		return nil
	}); err != nil {
		return found, err
	}
	return found, nil
}

func (f Options) included(path string) bool {
	if !f.extensionIncluded(filepath.Ext(path)) {
		return false
	}

	if len(f.Include) > 0 {
		for _, pattern := range f.Include {
			if ok, err := doublestar.Match(pattern, path); err == nil && ok {
				return true
			}
		}
		return false
	}

	return true
}

func (f Options) excluded(path string) bool {
	if len(f.Exclude) == 0 {
		return false
	}

	for _, pattern := range f.Exclude {
		if ok, err := doublestar.Match(pattern, path); err == nil && ok {
			return true
		}
	}

	return false
}

func (f Options) extensionIncluded(ext string) bool {
	for _, e := range f.Extensions {
		if e == ext {
			return true
		}
	}
	return false
}
