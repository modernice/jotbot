package fixture

import "errors"

type Foo struct{}

// Foo is a method.
func (f *Foo) Foo() error {
	return f.foo()
}

// foo is a method.
func (f *Foo) foo() error {
	return f.bar()
}

// bar is a method, too.
func (f *Foo) bar() error {
	return errors.New("bar")
}
