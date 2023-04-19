package fixture

import "errors"

func Bar() error {
	return errors.New("bar")
}
