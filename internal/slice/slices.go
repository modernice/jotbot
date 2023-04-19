package slice

// Map applies a function [fn] to each element of a slice [s] and returns a new 
// slice of the resulting values. The input slice [s] can be of any type [In], 
// and the output slice will be of type [Out].
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

// Filter returns a new slice containing all elements of the input slice s that 
// satisfy the predicate function fn. The input slice s and the output slice 
// have the same type.
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

// Unique returns a new slice containing only the unique elements of the input 
// slice. The input slice must be comparable.
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
// [S]. The input slice [S] must be a slice of a comparable type [T]. The 
// returned slice [S] has the same type as the input slice [S]. If the input 
// slice [S] is nil, NoZero returns nil.
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
