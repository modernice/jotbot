package generate

import (
	"github.com/modernice/jotbot/generate"
)

var _ generate.Service = (*Service)(nil)

type Service struct {
	docs map[string]string // map[identifier]doc
}

func MockService() *Service {
	return &Service{
		docs: make(map[string]string),
	}
}

func (svc *Service) WithDoc(identifier, doc string) *Service {
	svc.docs[identifier] = doc
	return svc
}

func (svc *Service) GenerateDoc(ctx generate.Context) (string, error) {
	id := ctx.Identifier()

	doc, ok := svc.docs[id]
	if !ok {
		return "", nil
	}

	return doc, nil
}
