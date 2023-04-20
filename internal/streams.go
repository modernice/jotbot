package internal

// Drain[T any](vals <-chan T, errs <-chan error) ([]T, error)
//
// Drain reads values from a channel of type T and returns a slice of those
// values. It also accepts an error channel and returns any error encountered
// while reading the channel. If an error is encountered, Drain stops reading
// from the value channel and returns the accumulated slice along with the
// error.
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

// Walk iterates over values received from a channel of type T, and calls the
// function fn for each value. If fn returns an error, Walk immediately returns
// that error. Walk continues until the channel is closed or fn returns an
// error.
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
