package patch

import (
	"context"

	"github.com/spf13/afero"
)

type MockPatch[Content ~string | ~[]byte] struct {
	files map[string]Content
}

func Mock[Content interface{ ~string | ~[]byte }](files map[string]Content) *MockPatch[Content] {
	return &MockPatch[Content]{files}
}

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
