package http

import "TLExtractor/http/types"

type multiPartData map[string]string

func (ct multiPartData) Apply(o *types.MultiPartInfo) {
	o.Data = ct
}

func Data(data map[string]string) MultiPartOption {
	return multiPartData(data)
}
