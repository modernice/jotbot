package generate

import (
	"fmt"

	"github.com/modernice/opendocs/generate"
)

var _ generate.Service = (*Service)(nil)

// Service is a type that implements the generate.Service interface. It
// represents a service that generates documentation for Go code. It has a map
// of maps that stores documentation for each file and identifier. It provides
// methods to add documentation for a file and identifier, generate
// documentation for a given context, and get all generations of documentation.
type Service struct {
	Fallbacks bool
	docs      map[string]map[string]string // map[file]map[identifier]doc
}

// MockService is a function that returns a pointer to a Service. Service is a
// struct that implements the generate.Service interface. It has a map of maps
// that stores documentation for each identifier in each file. MockService
// returns an empty Service with an empty map of maps.
func MockService() *Service {
	return &Service{
		docs: make(map[string]map[string]string),
	}
}

// Generations returns a slice of
// [generate.Generation](https://pkg.go.dev/github.com/modernice/opendocs/generate#Generation)
// values, each representing a documented identifier in a file.
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

// WithDoc adds documentation for an identifier in a file to the Service. It
// takes in three string arguments: the path of the file, the identifier, and
// the documentation. It returns a pointer to the Service.
func (svc *Service) WithDoc(path, identifier, doc string) *Service {
	if _, ok := svc.docs[path]; !ok {
		svc.docs[path] = make(map[string]string)
	}
	svc.docs[path][identifier] = doc
	return svc
}

// GenerateDoc returns the documentation for a given file and identifier. It
// takes a generate.Context as input and returns a string and an error. If the
// documentation is not found, it returns an empty string and an error.
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
