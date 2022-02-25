package box

// MaxI64 returns max(a,b)
func MaxI64(a, b int64) int64 {
	if a >= b {
		return a
	}

	return b
}

// MaxI32 returns max(a,b)
func MaxI32(a, b int32) int32 {
	if a > b {
		return a
	}

	return b
}

// MaxInt returns max(a,b)
func MaxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// MaxU64 returns max(a,b)
func MaxU64(a, b uint64) uint64 {
	if a >= b {
		return a
	}

	return b
}

// MaxU32 returns max(a,b)
func MaxU32(a, b uint32) uint32 {
	if a > b {
		return a
	}

	return b
}

// MaxUInt returns max(a,b)
func MaxUInt(a, b uint) uint {
	if a > b {
		return a
	}

	return b
}

// Clamp returns the value rounded to [min, max]
func ClampI32(min, max, val int32) int32 {
	if val < min {
		return min
	}

	if val > max {
		return max
	}

	return val
}

// AbsI64 returns |x|
func AbsI64(v int64) uint64 {
	if v > 0 {
		return uint64(v)
	}

	return uint64(-v)
}
