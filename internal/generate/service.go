package generate

import (
	"fmt"

	"github.com/modernice/opendocs/generate"
)

var _ generate.Service = (*Service)(nil)

type Service struct {
	docs map[string]map[string]string // map[file]map[identifier]doc
}

func MockService() *Service {
	return &Service{
		docs: make(map[string]map[string]string),
	}
}

func (svc *Service) Generations() []generate.Generation {
	generations := make([]generate.Generation, 0, len(svc.docs))
	for file, identifiers := range svc.docs {
		for identifier, doc := range identifiers {
			generations = append(generations, generate.Generation{
				Path:       file,
				Identifier: identifier,
				Doc:        doc,
			})
		}
	}
	return generations
}

func (svc *Service) WithDoc(path, identifier, doc string) *Service {
	if _, ok := svc.docs[path]; !ok {
		svc.docs[path] = make(map[string]string)
	}
	svc.docs[path][identifier] = doc
	return svc
}

func (svc *Service) GenerateDoc(ctx generate.Context) (string, error) {
	file := ctx.File()
	id := ctx.Identifier()

	docs, ok := svc.docs[file]
	if !ok {
		return "", fmt.Errorf("no docs for file %q", file)
	}

	doc, ok := docs[id]
	if !ok {
		return "", fmt.Errorf("no docs for identifier %q in file %q", id, file)
	}

	return doc, nil
}
