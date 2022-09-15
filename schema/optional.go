package schema

// PointerTo turns a value into a pointer to that value. This makes it easy to pass when a pointer is expected.
func PointerTo[T any](value T) *T {
	return &value
}

// Int creates an int64 out of the specified int-like type.
func Int[T ~int](value T) int64 {
	return int64(value)
}

// IntPointer creates a pointer to an int after converting it to int64.
func IntPointer[T ~int](value T) *int64 {
	v := int64(value)
	return &v
}
