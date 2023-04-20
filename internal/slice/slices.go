package slice

// Map applies a given function to each element of a slice[[]], returning a new 
// slice[] with the results. The function takes an element of the input slice as 
// its argument and returns the corresponding element of the output slice[].
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

// Filter [Filter[T]] is a function that takes a slice of type T and a function 
// that takes a value of type T and returns a boolean. It returns a new slice 
// containing only the elements of the input slice for which the function 
// returns true.
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

// Unique returns a new slice with only the unique elements of the input slice. 
// The function accepts a slice of type [S ~[]T, T comparable](S), where T is a 
// comparable type. The resulting slice has the same underlying type as the 
// input slice.
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

// NoZero removes all zero values from a given slice [S] of comparable type [T]. 
// It returns a new slice with only the non-zero values.
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
