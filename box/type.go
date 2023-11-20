package box

func IsZero[T comparable](v T) bool {
	return v == *new(T)
}
