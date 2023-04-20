package slice

// Map applies a function to each element of a slice and returns a new slice
// containing the transformed elements. The input slice can contain any type of
// element, and the output slice can have a different type. The function to be
// applied must take an input element and return an output element of a possibly
// different type.
func Map[In, Out any](s []In, fn func(In) Out) []Out {
	if s == nil {
		return nil
	}

	out := make([]Out, len(s))
	for i, v := range s {
		out[i] = fn(v)
	}
	return out
}

// Filter returns a new slice containing all elements of the input slice that
// satisfy the provided boolean function.
func Filter[S ~[]T, T any](s S, fn func(T) bool) S {
	if s == nil {
		return nil
	}

	out := make(S, 0, len(s))
	for _, v := range s {
		if fn(v) {
			out = append(out, v)
		}
	}
	return out
}

// Unique returns a new slice containing all the unique elements of the input
// slice, in the order in which they first appear. The input slice must be
// sorted or comparable.
func Unique[S ~[]T, T comparable](s S) S {
	if s == nil {
		return nil
	}

	found := make(map[T]struct{}, len(s))
	out := make(S, 0, len(s))
	for _, v := range s {
		if _, ok := found[v]; !ok {
			out = append(out, v)
			found[v] = struct{}{}
		}
	}
	return out
}

// NoZero returns a new slice with all zero values removed from the input slice
// [S]. The input slice must be comparable.
func NoZero[S ~[]T, T comparable](s S) S {
	if s == nil {
		return nil
	}

	var zero T
	out := make(S, 0, len(s))
	for _, v := range s {
		if v != zero {
			out = append(out, v)
		}
	}
	return out
}
