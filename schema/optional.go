package schema

// Optional turns a value into a pointer to that value. This makes it easy to pass when a pointer is expected.
func Optional[T any](value T) *T {
	return &value
}
