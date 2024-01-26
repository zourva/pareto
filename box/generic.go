package box

import "fmt"

type Number interface {
	int | uint | int32 | uint32 | int64 | uint64 | float32 | float64
}

type Integer interface {
	int | uint | int32 | uint32 | int64 | uint64
}

type Float interface {
	float32 | float64
}

func Abs[T Number](t T) T {
	if t < 0 {
		return -t
	}

	return t
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

func FEqual[T Number](a, b, epsilon T) bool {
	return Abs(a-b) <= epsilon
}

// FloatToStr float32 and float64 are expected.
func FloatToStr[T Float](t T, precision int) string {
	if precision < 0 {
		precision = 6
	}

	ff := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(ff, t)
}

// IntToStr int/uint/int32/uint32/int64/uint64
// are expected.
func IntToStr[T Integer](t T) string {
	return fmt.Sprintf("%d", t)
}

//func Max[T Number](a, b T) T {
//	return T(math.Max(float64(a), float64(b)))
//}

// SetIfEq sets value of v to def if v == zero.
func SetIfEq[T Integer](v *T, zero, def T) {
	if *v == zero {
		*v = def
	}
}
