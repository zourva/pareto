package box

type Number interface {
	int | uint | int32 | uint32 | int64 | uint64 | float32 | float64
}

//func Max[T Number](a, b T) T {
//	return T(math.Max(float64(a), float64(b)))
//}

func Abs[T Number](t T) T {
	if t < 0 {
		return -t
	}

	return t
}

func FEqual[T Number](a, b, epsilon T) bool {
	return Abs(a-b) <= epsilon
}

func Clamp[T Number](val *T, min, max T) T {
	if *val < min {
		*val = min
		return min
	}

	if *val > max {
		*val = max
		return max
	}

	return *val
}

func ClampDefault[T Number](val *T, min, max, def T) T {
	if *val < min {
		*val = def
		return min
	}

	if *val > max {
		*val = def
		return max
	}

	return *val
}
