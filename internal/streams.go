package internal

func Drain[T any](vals <-chan T, errs <-chan error) ([]T, error) {
	out := make([]T, 0, len(vals))
	for {
		select {
		case err, ok := <-errs:
			if !ok {
				return out, nil
			}
			return out, err
		case v, ok := <-vals:
			if !ok {
				return out, nil
			}
			out = append(out, v)
		}
	}
}
