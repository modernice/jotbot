package ts

import (
	"context"
	"fmt"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
)

var (
	FileExtensions = []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"}
)

type Service struct {
	finder *Finder
}

type Option func(*Service)

func WithFinder(f *Finder) Option {
	return func(s *Service) {
		s.finder = f
	}
}

func New(opts ...Option) *Service {
	var svc Service
	for _, opt := range opts {
		opt(&svc)
	}
	if svc.finder == nil {
		svc.finder = NewFinder()
	}
	return &svc
}

func (svc *Service) Extensions() []string {
	return FileExtensions
}

func (svc *Service) Find(code []byte) ([]find.Finding, error) {
	return svc.finder.Find(context.Background(), code)
}

func (svc *Service) Prompt(input generate.Input) string {
	return Prompt(input)
}

func (svc *Service) Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error) {
	pos, err := svc.finder.Position(ctx, identifier, code)
	if err != nil {
		return nil, fmt.Errorf("find position of %q in code: %w", identifier, err)
	}
	return InsertComment(doc, code, pos)
}
