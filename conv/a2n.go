package conv

import (
	"math"
	"strconv"
)

// Atoi tries to parse a string into an integer of int64 type,
// and returns math.MaxInt64 when conversion fails.
func Atoi(n string) int64 {
	r, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return math.MaxInt64
	}

	return r
}
