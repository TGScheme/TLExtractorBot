package http

import (
	"TLExtractor/http/types"
	"time"
)

type timeoutOption time.Duration

func (ct timeoutOption) Apply(o *types.RequestOptions) {
	o.Timeout = time.Duration(ct)
}

func Timeout(time time.Duration) RequestOption {
	return timeoutOption(time)
}
