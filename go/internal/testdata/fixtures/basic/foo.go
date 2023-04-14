package repo

import "errors"

func Foo() error {
	return errors.New("foo")
}
