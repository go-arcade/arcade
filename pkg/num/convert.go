package num

import (
	"fmt"
	"math"
)

// MustInt converts uint64 to int, panics if overflow
func MustInt(val uint64) int {
	if val > math.MaxInt {
		panic(fmt.Sprintf("numutil: uint64 value %d overflows int", val))
	}
	return int(val)
}

// MustInt64 converts uint64 to int64, panics if overflow
func MustInt64(val uint64) int64 {
	if val > math.MaxInt64 {
		panic(fmt.Sprintf("numutil: uint64 value %d overflows int64", val))
	}
	return int64(val)
}

// MustUint8 converts int to uint8, panics if overflow or negative
func MustUint8(val int) uint8 {
	if val < 0 {
		panic(fmt.Sprintf("numutil: int value %d is negative, cannot convert to uint8", val))
	}
	if val > math.MaxUint8 {
		panic(fmt.Sprintf("numutil: int value %d overflows uint8", val))
	}
	return uint8(val)
}

// MustUint64 converts int64 to uint64, panics if negative
func MustUint64(val int64) uint64 {
	if val < 0 {
		panic(fmt.Sprintf("numutil: int64 value %d is negative, cannot convert to uint64", val))
	}
	return uint64(val)
}
