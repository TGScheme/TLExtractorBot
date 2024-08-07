package java

import (
	"strings"
)

func ParseType(src string) string {
	switch src {
	case "Int32":
		return "int"
	case "Int64":
		return "long"
	case "Double", "Bool":
		return strings.ToLower(src)
	case "NativeByteBuffer", "ByteArray", "ByteBuffer", "Bytes":
		return "bytes"
	}
	return src
}
