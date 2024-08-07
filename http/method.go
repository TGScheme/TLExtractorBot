package http

import "TLExtractor/http/types"

type methodOption string

func (ct methodOption) Apply(o *types.RequestOptions) {
	o.Method = string(ct)
}

func Method(method string) RequestOption {
	return methodOption(method)
}
