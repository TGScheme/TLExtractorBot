package http

import "TLExtractor/http/types"

type bodyOption []byte

func (ct bodyOption) Apply(o *types.RequestOptions) {
	o.Body = ct
}

func Body(body []byte) RequestOption {
	return bodyOption(body)
}
