# opendocs

## Usage

```go
package example

func example(root string) {
	docs := opendocs.Repo(root)

	// generate, then commit
	patch, err := docs.Patch() // generate patch
	if err != nil {
		panic(err)
	}
	err := git.Repo("/path/to/repo").Commit(patch)

	// or both in one
	patch, err := docs.Generate()
}
```
