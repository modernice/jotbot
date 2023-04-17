package fixture

type X struct{}

func (X) Foo() string {
	return "foo"
}

func (*X) Bar() string {
	return "bar"
}

// Y already has a description, should not be found by finder.
type Y struct{}

func (Y) Foo() string {
	return "foo"
}

// Bar should not be found by finder.
func (Y) Bar() string {
	return "bar"
}
