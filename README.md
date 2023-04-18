# opendocs

`opendocs` writes documentation for your Go code (support for other languages
is planned). Documentation is generated using OpenAI's GPT-3 (or GPT-3.5-Turbo) API.

## Requirements

Git is not required to use `opendocs`. However, if you want to use the automatic
commit feature, you need to have Git installed on your machine.

## Installation

```sh
go get github.com/modernice/opendocs
```

## Usage

### Library

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modernice/opendocs"
	"github.com/modernice/opendocs/services/openai"
)

func main() {
	root := "/path/to/your/go/project"
	svc := openai.New("YOUR-API-KEY")
	repo := opendocs.Repo(root)

	result, err := repo.Generate(ctx, svc)
	// handle err

	err = result.Commit(root) // requires Git
	// handle err

	err = result.Patch().Apply(root) // does not require Git
	// handle err
}
```

### Configurable Finder

```go
package example

func main() {
	f := find.New(
		// pre-defined skips
		find.Skip{
			Hidden: true,
			Testdata: true,
			Testfiles: true,
		},

		// custom skip function
		find.SkipFunc(d fs.DirEntry) bool {
			return false
		},
	)

	f.Uncommented()
}
```
