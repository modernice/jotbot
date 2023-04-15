# opendocs

## Usage

### Basic

```go
package example

func example(root string) {
	repo := opendocs.Repo(root)

	// first generate the patch
	patch, err := docs.Patch(context.TODO()) // generate patch
	if err != nil {
		panic(err)
	}
	
	// then commit to the repo
	err := git.Repo("/path/to/repo").Commit(patch)

	// or all in one
	patch, err := docs.Generate(context.TODO())
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
