package io

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
)

func Scanln(a any) error {
	scanner := bufio.NewScanner(os.Stdin)
	var tmp string
	if scanner.Scan() {
		tmp = scanner.Text()
	}
	switch a.(type) {
	case *string:
		*a.(*string) = tmp
	case *int, *int64, *int32, *int16, *int8:
		i, err := strconv.Atoi(tmp)
		if err != nil {
			return fmt.Errorf("could not convert %s to int: %s", tmp, err)
		}
		switch a.(type) {
		case *int:
			*a.(*int) = i
		case *int64:
			*a.(*int64) = int64(i)
		case *int32:
			*a.(*int32) = int32(i)
		case *int16:
			*a.(*int16) = int16(i)
		case *int8:
			*a.(*int8) = int8(i)
		}
	default:
		return errors.New("unsupported type")
	}
	return nil
}
