# opendocs

## Usage

```go
package example

func example(root string) {
	od := opendocs.New(root)

	patches, err := od.Generate() // generate patches
	if err != nil {
		panic(err)
	}

	for path, patch := range patches {
		path // path to file
		patch // *opendocs.Patcher
	}
}
```
