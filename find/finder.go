package find

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	// DefaultExtensions represents the set of file extensions that are searched for
	// by default.
	DefaultExtensions = []string{
		".go",
		".ts",
	}

	// DefaultExclude represents a list of glob patterns used to identify file paths
	// that should be omitted from search or processing operations. These patterns
	// are designed to match common directories and files that are typically not of
	// interest, such as hidden directories, distribution folders, dependency
	// directories, vendor folders, test-related files, and generated protocol
	// buffer code files.
	DefaultExclude = []string{
		"**/.*/**",
		"**/dist/**",
		"**/node_modules/**",
		"**/vendor/**",
		"**/testdata/**",
		"**/test/**",
		"**/tests/**",
		"**/*.pb.go",
	}

	// Default represents the standard configuration for file searching,
	// encompassing common file extensions to include and patterns to exclude.
	Default = Options{
		Extensions: DefaultExtensions,
		Exclude:    DefaultExclude,
	}
)

// Options represents a set of configurable parameters used to modify the
// behavior of file searching operations within a file system. It allows
// specifying file extensions to include, patterns to specifically include, and
// patterns to exclude during the search. The Options can be adjusted using
// provided functional options that set the appropriate fields for extensions,
// inclusion, and exclusion patterns. These settings are then applied when
// performing a file search to determine which files are considered matches
// based on the criteria defined by the Options instance.
type Options struct {
	Extensions []string
	Include    []string
	Exclude    []string
}

// Option represents a configuration modifier which applies custom settings to
// an Options object. It is used to specify inclusion and exclusion patterns, as
// well as file extensions that should be considered during the file search
// process. Option functions are intended to be passed to other functions in the
// package that require configurable search criteria, allowing users to tailor
// the behavior of file discovery according to their needs.
type Option func(*Options)

// Extensions sets the file extensions that should be used to filter files
// during the search process. It overrides any previously set extensions with
// the provided list of extensions. This is one of several options that can be
// applied to configure the behavior of a file search.
func Extensions(exts ...string) Option {
	return func(o *Options) {
		o.Extensions = exts
	}
}

// Include appends the provided file patterns to the list of patterns that will
// be included during the file search process, modifying the search criteria
// encapsulated within an [Options] instance.
func Include(patterns ...string) Option {
	return func(o *Options) {
		o.Include = append(o.Include, patterns...)
	}
}

// Exclude appends the given patterns to the list of patterns used to exclude
// files or directories in file search configurations. It returns an Option
// that, when applied to an Options object, modifies its Exclude field to
// include these additional patterns.
func Exclude(patterns ...string) Option {
	return func(o *Options) {
		o.Exclude = append(o.Exclude, patterns...)
	}
}

// Files searches for files within a given file system that match specified
// patterns, taking into account inclusion and exclusion criteria. It applies
// options to configure the search behavior, such as filtering by file
// extensions or specific file paths. The function returns a slice of file paths
// that meet the criteria along with any error encountered during the search
// process. The context parameter allows the search to be canceled or have a
// deadline.
func Files(ctx context.Context, files fs.FS, opts ...Option) ([]string, error) {
	cfg := Default
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg.Find(ctx, files)
}

// Find traverses the file system starting from the root directory, applying
// inclusion and exclusion patterns, and returns a slice of file paths that
// match the specified criteria within the given context. It respects the
// configured file extensions, inclusion, and exclusion patterns to determine
// which files are included in the results. If an error occurs during traversal,
// it returns the successfully found files up to that point along with the
// encountered error.
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
