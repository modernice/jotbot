package fixture

func Foo() {}

type X[T any] struct{}

func (x *X[_]) Foo() {}

type y[T any] struct{}

func (y *y[T]) Foo() {}
