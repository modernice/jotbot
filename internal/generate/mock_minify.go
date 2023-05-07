package generate

// MockMinifier is a mock implementation of a minifier that returns the provided
// byte slice as-is, without performing any actual minification. It is useful
// for testing purposes.
type MockMinifier []byte

// Minify returns a minified version of the given byte slice using the
// MockMinifier receiver.
func (m MockMinifier) Minify([]byte) ([]byte, error) {
	return []byte(m), nil
}
