# opendocs

## Usage

### Basic

```go
package example

func example(root string) {
	repo := opendocs.New(root)

	// first generate the docs
	result, err := repo.Generate(context.TODO()) // generate docs
	if err != nil {
		panic(err)
	}
	
	// then commit to the repo
	err := result.Commit("/path/to/repo")

	// or without git
	patch := result.Patch()
	patch.Apply("/path/to/repo")
}
```

### Generate docs

```go
package example

func example(repo fs.FS) {
	var svc generate.Service // can be openai but also anything else

	g := generate.New(svc)

	result, err := g.Generate(context.TODO(), repo) // map[string][]Generation

	for path, generations := range result {
		for _, gen := range generations {
			gen.Identifier // identifier of the type or function
			gen.Doc // the generated doc
		}
	}

	p, err := result.Patch() // generate a patch from the result
}
```
