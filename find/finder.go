package find

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

var (
	DefaultExtensions = []string{
		".go",
		".ts",
	}

	DefaultExclude = []string{
		"**/.*/**", // hidden directories
	}

	Default = Options{
		Extensions: DefaultExtensions,
		Exclude:    DefaultExclude,
	}
)

type Options struct {
	Extensions []string
	Include    []string
	Exclude    []string
}

type Option func(*Options)

func Extensions(exts ...string) Option {
	return func(o *Options) {
		o.Extensions = exts
	}
}

func Include(patterns ...string) Option {
	return func(o *Options) {
		o.Include = append(o.Include, patterns...)
	}
}

func Exclude(patterns ...string) Option {
	return func(o *Options) {
		o.Exclude = append(o.Exclude, patterns...)
	}
}

func Files(ctx context.Context, files fs.FS, opts ...Option) ([]string, error) {
	cfg := Default
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg.Find(ctx, files)
}

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
			return nil
		}

		if !f.included(path) {
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
		var included bool
		for _, pattern := range f.Include {
			if ok, err := doublestar.Match(pattern, path); err == nil && ok {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	if len(f.Exclude) > 0 {
		for _, pattern := range f.Exclude {
			if ok, err := doublestar.Match(pattern, path); err == nil && ok {
				return false
			}
		}
	}

	return true
}

func (f Options) extensionIncluded(ext string) bool {
	for _, e := range f.Extensions {
		if e == ext {
			return true
		}
	}
	return false
}
