package http

import "TLExtractor/http/types"

type multiPartOption types.MultiPartInfo

func (ct multiPartOption) Apply(o *types.RequestOptions) {
	tmpMultiPartInfo := types.MultiPartInfo(ct)
	o.MultiPart = &tmpMultiPartInfo
}

func MultiPartForm(options ...MultiPartOption) RequestOption {
	var multiPartInfo types.MultiPartInfo
	for _, option := range options {
		option.Apply(&multiPartInfo)
	}
	return multiPartOption(multiPartInfo)
}
