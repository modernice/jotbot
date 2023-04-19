package bar

import "errors"

func Foo() error {
	return errors.New("foo")
}
