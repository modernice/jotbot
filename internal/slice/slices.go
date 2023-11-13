package slice

// Map applies a provided function to each element of the given slice, returning
// a new slice with the results.
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

// Filter returns a new slice holding only the elements of s that satisfy fn.
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

// Unique removes duplicate elements from the provided slice, preserving the
// order of the first occurrence of each element. It returns a slice containing
// only the unique elements.
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

// NoZero removes zero values from the provided slice and returns a new slice
// containing only non-zero values of the same type. It maintains the order of
// the original elements.
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
