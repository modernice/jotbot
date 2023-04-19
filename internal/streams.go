package internal

// Drain reads values from a channel of type T and returns a slice of type T 
// containing all the values read from the channel. It also reads from a channel 
// of errors and returns the first error encountered.
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

// Walk iterates over values received from a channel [vals] and applies a 
// function [fn] to each value. If the function returns an error, Walk stops 
// iteration and returns the error. If the channel [errs] is closed, Walk stops 
// iteration and returns nil.
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
