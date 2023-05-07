package patch

import (
	"context"

	"github.com/spf13/afero"
)

// MockPatch is a type that represents a collection of files with their
// contents, used to apply these files to a specified root directory within the
// filesystem. It supports both string and byte slice content types.
type MockPatch[Content ~string | ~[]byte] struct {
	files map[string]Content
}

// Mock creates a new MockPatch instance with the provided map of files and
// their corresponding content.
func Mock[Content interface{ ~string | ~[]byte }](files map[string]Content) *MockPatch[Content] {
	return &MockPatch[Content]{files}
}

// Apply writes the contents of the MockPatch to the specified root directory
// using the provided context. It returns an error if any issues occur during
// file creation or writing.
func (p *MockPatch[_]) Apply(ctx context.Context, root string) error {
	fs := afero.NewBasePathFs(afero.NewOsFs(), root)

	for path, content := range p.files {
		f, err := fs.Create(path)
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(content)); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
