package schema

// PointerTo turns a value into a pointer to that value. This makes it easy to pass when a pointer is expected.
func PointerTo[T any](value T) *T {
	return &value
}
