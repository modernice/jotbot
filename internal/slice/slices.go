package slice

func Map[In, Out any](s []In, fn func(In) Out) []Out {
	out := make([]Out, len(s))
	for i, v := range s {
		out[i] = fn(v)
	}
	return out
}
