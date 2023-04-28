package golang

import (
	"context"

	"github.com/modernice/jotbot/find"
)

type Service struct {
	finder *Finder
}

func New(finder *Finder) *Service {
	return &Service{
		finder: finder,
	}
}

func (svc *Service) Find(ctx context.Context, code []byte) ([]find.Finding, error) {
	return svc.finder.Find(ctx, code)
}
