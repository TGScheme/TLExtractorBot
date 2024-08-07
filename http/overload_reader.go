package http

import (
	"TLExtractor/http/types"
	"io"
)

type overloadReader func(r io.Reader) io.Reader

func (ct overloadReader) Apply(o *types.RequestOptions) {
	o.OverloadReader = ct
}

func OverloadReader(reader overloadReader) RequestOption {
	return reader
}
