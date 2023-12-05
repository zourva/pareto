package box

import "fmt"

type Number interface {
	int | uint | int32 | uint32 | int64 | uint64 | float32 | float64
}

// IntToStr int/uint/int32/uint32/int64/uint64
// are expected.
func IntToStr[T int | uint | int32 | uint32 | int64 | uint64](t T) string {
	return fmt.Sprintf("%d", t)
}

// FloatToStr float32 and float64 are expected.
func FloatToStr[T float32 | float64](t T, precision int) string {
	if precision < 0 {
		precision = 6
	}

	ff := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(ff, t)
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
