package fixture

type X struct{}

func (X) Foo() {}

type Y struct{}

func (*Y) Foo() {}

func Foo() {}
