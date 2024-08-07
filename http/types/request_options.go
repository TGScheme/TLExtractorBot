package types

import (
	"io"
)

type RequestOptions struct {
	Method         string
	Body           []byte
	Headers        map[string]string
	MultiPart      *MultiPartInfo
	OverloadReader func(r io.Reader) io.Reader
}
