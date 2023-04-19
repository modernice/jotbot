package slice

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
