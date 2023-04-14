# opendocs

## Usage

```go
package example

func example(root string) {
	od := opendocs.New(root)

	patch, err := od.Generate() // generate patches
	if err != nil {
		panic(err)
	}
}
```
