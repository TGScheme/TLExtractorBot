package scheme

import (
	"fmt"
	"strconv"
)

func ParseConstructor(constructor string) string {
	res, _ := strconv.Atoi(constructor)
	return fmt.Sprintf("%x", uint32(res))
}
