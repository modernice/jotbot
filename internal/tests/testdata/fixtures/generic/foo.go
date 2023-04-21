package fixture

func Foo() {}

type X[T, T2 any] struct{}

func (x *X[T, T2]) Foo() {}

type y[T, T2 any] struct{}

func (y *y[T, _]) Foo() {}
