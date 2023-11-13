package internal

// Stream initiates a streaming of the provided values through a channel of the
// same type, allowing for concurrent processing of the values in a non-blocking
// manner. It returns a receive-only channel from which the streamed values can
// be read.
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

// Drain collects values from a channel of a specified type until the channel is
// closed, while also listening for an error on a separate error channel. It
// returns a slice containing all the collected values and an error if
// encountered. If the error channel is closed without any errors sent, it
// returns the collected values and a nil error. If an error is received, it
// returns the values collected up to that point along with the received error.
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

// Walk applies a provided function to each element received from a channel,
// halting on and returning the first error encountered. It consumes two
// channels: one for values of type T and another for errors.
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

// Map applies a provided function to each element received from a channel and
// returns a new channel that emits the results.
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
