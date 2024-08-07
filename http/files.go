package http

import "TLExtractor/http/types"

type multiPartFiles map[string]types.FileDescriptor

func (ct multiPartFiles) Apply(o *types.MultiPartInfo) {
	o.Files = ct
}

func Files(file map[string]types.FileDescriptor) MultiPartOption {
	return multiPartFiles(file)
}
