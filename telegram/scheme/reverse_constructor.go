package scheme

import (
	"strconv"
)

func ReverseConstructor(hexString string) string {
	value, _ := strconv.ParseUint(hexString, 16, 32)
	bits := len(hexString) * 4
	if bits == 32 && value&(1<<31) != 0 {
		value -= 1 << 32
	}
	return strconv.FormatInt(int64(value), 10)
}
