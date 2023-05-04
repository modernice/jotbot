package generate

type MockMinifier []byte

func (m MockMinifier) Minify([]byte) ([]byte, error) {
	return []byte(m), nil
}
