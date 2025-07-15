package optional

// Optional uses implementation
// from https://github.com/frenchie4111/go-generic-optional/blob/main/opt.go
type Optional[T any] struct {
	value  T
	exists bool
}

// Empty creates a new Optional without a value.
func Empty[T any]() Optional[T] {
	var empty T
	return Optional[T]{empty, false}
}

// Of creates a new Optional with a value.
func Of[T any](value T) Optional[T] {
	return Optional[T]{value, true}
}

// Exists returns true if the optional value has been set
func (o Optional[T]) Exists() bool {
	return o.exists
}

// Get returns the value.
// It's invalid to use the returned value if the bool is false.
func (o Optional[T]) Get() T {
	return o.value
}

// GetOrElse returns the value if it exists and returns defaultValue otherwise.
func (o Optional[Value]) GetOrElse(defaultValue Value) Value {
	if !o.exists {
		return defaultValue
	}

	return o.value
}
