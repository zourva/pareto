package box

import (
	"math"
	"strconv"
)

// I32toa returns the string presentation of an integer of int32 type.
func I32toa(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

// U32toa returns the string presentation of an integer of uint32 type.
func U32toa(n uint32) string {
	return strconv.FormatUint(uint64(n), 10)
}

// U64toa returns the string presentation of an integer of uint64 type.
func U64toa(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// I64toa returns the string presentation of an integer of int64 type.
func I64toa(n int64) string {
	return strconv.FormatInt(n, 10)
}

// F64toa returns the string presentation of a float64 value.
func F64toa(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// F32toa returns the string presentation of a float32 value.
func F32toa(n float32) string {
	return strconv.FormatFloat(float64(n), 'f', -1, 32)
}

// Itoa returns the string presentation of an integer of int type.
func Itoa(n int) string {
	return strconv.FormatInt(int64(n), 10)
}

// Atoi tries to parse a string into an integer of int64 type,
// and returns math.MaxInt64 when conversion fails.
func Atoi(n string) int64 {
	r, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return math.MaxInt64
	}

	return r
}
