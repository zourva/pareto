package box

import (
	"math"
	"strconv"
)

func I32toa(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

func U32toa(n uint32) string {
	return strconv.FormatUint(uint64(n), 10)
}

func U64toa(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func I64toa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func F64toa(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func F32toa(n float32) string {
	return strconv.FormatFloat(float64(n), 'f', -1, 32)
}

func Itoa(n int) string {
	return strconv.FormatInt(int64(n), 10)
}

func Atoi(n string) int64 {
	r, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return math.MaxInt64
	}

	return r
}
