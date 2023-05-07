package internal

// Stream creates a new channel of type T, sends the provided values to the
// channel concurrently, and returns the channel. The channel is closed after
// all values have been sent.
func Stream[T any](values ...T) <-chan T {
	ch := make(chan T, len(values))
	go func() {
		defer close(ch)
		for _, v := range values {
			ch <- v
		}
	}()
	return ch
}

// Drain[T any](vals <-chan T, errs <-chan error) ([]T, error) function reads
// from the channel vals until it is closed and stores the read values into a
// slice of type T. If errs channel is closed, the function returns the stored
// values and nil error. If an error is received from errs channel, the function
// returns the stored values and that error.
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
// function [fn] to each value. If the function returns an error, Walk returns
// that error. If values are received on the error channel [errs], Walk stops
// iterating and returns the first error encountered.
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

// Map applies a given function [fn] to each value received from an input
// channel [in] of type In and sends the transformed values of type Out to an
// output channel. The output channel is closed when all values have been
// processed.
func Map[In, Out any](in <-chan In, fn func(In) Out) <-chan Out {
	out := make(chan Out)
	go func() {
		defer close(out)
		for v := range in {
			out <- fn(v)
		}
	}()
	return out
}
