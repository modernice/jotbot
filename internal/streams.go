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

func Walk[T any](vals <-chan T, errs <-chan error, fn func(T) error) error {
	for {
		select {
		case err, ok := <-errs:
			if !ok {
				return nil
			}
			return err
		case v, ok := <-vals:
			if !ok {
				return nil
			}
			if err := fn(v); err != nil {
				return err
			}
		}
	}
}
