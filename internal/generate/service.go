package generate

import (
	"fmt"

	"github.com/modernice/jotbot/generate"
)

var _ generate.Service = (*Service)(nil)

// Service [generate.Service] is a type that represents a service for generating
// documentation. It implements the generate.Service interface. Service stores
// documentation for code identifiers in each file and can generate
// documentation for specific identifiers. It has methods to add documentation
// to a file, get all generations of stored documentation, and generate
// documentation for a given identifier.
type Service struct {
	Fallbacks bool
	docs      map[string]map[string]string // map[file]map[identifier]doc
}

// MockService is a type that implements the generate.Service interface. It
// provides functionality for generating documentation for Go code. This type
// can be used to mock the Service interface and generate documentation in
// tests. Additionally, it provides methods to add documentation for a given
// file and identifier, retrieve all generations of documentation, and generate
// documentation for a given context.
func MockService() *Service {
	return &Service{
		docs: make(map[string]map[string]string),
	}
}

// Generations returns a slice of
// [generate.Generation](https://pkg.go.dev/github.com/modernice/jotbot/generate#Generation)
// that represents all the documented identifiers in the Service. Each
// Generation has its file path, identifier name, and its corresponding
// documentation.
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

// WithDoc sets the documentation for a given identifier in a specific file. It
// returns a pointer to the Service to allow for method chaining. This method
// belongs to the generate.Service interface [generate.Service].
func (svc *Service) WithDoc(path, identifier, doc string) *Service {
	if _, ok := svc.docs[path]; !ok {
		svc.docs[path] = make(map[string]string)
	}
	svc.docs[path][identifier] = doc
	return svc
}

// GenerateDoc returns the documentation for the identifier specified in a given
// file. It takes a generate.Context as an argument and returns the
// documentation as a string and an error. If there is no documentation for the
// identifier or file, it will return an empty string and an error. If Fallbacks
// is set to true, it will return an empty string instead of an error for
// missing documentation. This function implements the Service interface
// [generate.Service].
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
