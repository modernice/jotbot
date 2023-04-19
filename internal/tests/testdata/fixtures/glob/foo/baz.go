package foo

import "errors"

func Baz() error {
	return errors.New("baz")
}
