package optional

// from https://github.com/frenchie4111/go-generic-optional/blob/main/opt.go
type Optional[T any] struct {
	value  T
	exists bool
}

func Empty[T any]() Optional[T] {
	var empty T

	return Optional[T]{empty, false}
}

func Of[T any](value T) Optional[T] {
	return Optional[T]{value, true}
}

func (o Optional[T]) Exists() bool {
	return o.exists
}

func (o Optional[T]) Get() T {
	return o.value
}

func (o Optional[Value]) GetOrElse(defaultValue Value) Value {
	if !o.exists {
		return defaultValue
	}

	return o.value
}
