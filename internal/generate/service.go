package generate

import (
	"fmt"

	"github.com/modernice/jotbot/generate"
)

var _ generate.Service = (*Service)(nil)

// Service represents a service that generates documentation for code. It
// implements the generate.Service interface. It contains a map of
// documentations for each file and identifier. Use WithDoc to add
// documentations to the service. Use Generations to return all generations of
// documentations. Use GenerateDoc to generate documentation for a given file
// and identifier in a given context.
type Service struct {
	Fallbacks bool
	docs      map[string]map[string]string // map[file]map[identifier]doc
}

// MockService provides a mock implementation of generate.Service. It allows you
// to generate documentation for Go source code files by providing the path,
// identifier, and documentation. You can use Generations to retrieve all the
// generations of the provided path and identifier. WithDoc allows you to add
// documentation for a given path and identifier. GenerateDoc generates the
// documentation for the given path and identifier.
func MockService() *Service {
	return &Service{
		docs: make(map[string]map[string]string),
	}
}

// Generations returns a slice of
// [generate.Generation](https://godoc.org/github.com/modernice/jotbot/generate#Generation)
// values. These describe the file, identifier, and documentation for each
// identifier registered with the Service.
func (svc *Service) Generations() []generate.Generation {
	generations := make([]generate.Generation, 0, len(svc.docs))
	for file, identifiers := range svc.docs {
		for identifier, doc := range identifiers {
			generations = append(generations, generate.Generation{
				File:       file,
				Identifier: identifier,
				Doc:        doc,
			})
		}
	}
	return generations
}

// WithDoc adds documentation for a given identifier in a file to the Service.
// It returns a pointer to the modified Service.
func (svc *Service) WithDoc(path, identifier, doc string) *Service {
	if _, ok := svc.docs[path]; !ok {
		svc.docs[path] = make(map[string]string)
	}
	svc.docs[path][identifier] = doc
	return svc
}

// GenerateDoc generates the documentation for a given identifier in a file
// using the internal storage of documentation in the *Service type.
func (svc *Service) GenerateDoc(ctx generate.Context) (string, error) {
	file := ctx.File()
	id := ctx.Identifier()

	docs, ok := svc.docs[file]
	if !ok {
		if svc.Fallbacks {
			return "", nil
		}
		return "", fmt.Errorf("no docs for file %q", file)
	}

	doc, ok := docs[id]
	if !ok {
		if svc.Fallbacks {
			return "", nil
		}
		return "", fmt.Errorf("no docs for identifier %q in file %q", id, file)
	}

	return doc, nil
}
