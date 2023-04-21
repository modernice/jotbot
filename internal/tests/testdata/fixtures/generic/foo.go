package fixture

func Foo() {}

type X[T any] struct{}

func (x *X[_]) Foo() {}
