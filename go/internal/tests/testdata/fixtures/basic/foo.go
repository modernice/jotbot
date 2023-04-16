package fixture

import "errors"

func Foo() error {
	return errors.New("foo")
}
