package http

import "TLExtractor/http/types"

type retriesOption int

func (ct retriesOption) Apply(o *types.RequestOptions) {
	o.Retries = int(ct)
}

func Retries(count int) RequestOption {
	return retriesOption(count)
}
