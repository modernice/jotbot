# git integration

## Usage

```go
package example

func example(fs fs.FS) {
	repo := New(fs)

	p := opendocs.NewPatcher(strings.NewReader(`...`))

	if err := repo.Commit(p); err != nil {
		panic(err)
	}
}
```
