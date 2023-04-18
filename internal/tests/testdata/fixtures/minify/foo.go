package fixture

import "errors"

type Foo struct{}

func (f *Foo) Foo() error {
	return f.foo()
}

func (f *Foo) foo() error {
	return f.bar()
}

func (f *Foo) bar() error {
	return errors.New("bar")
}
