package types

type TLObjDifferences struct {
	Object        TLInterface
	IsNew         bool
	IsDeleted     bool
	ChangedResult *TlDifferentResult
	NewFields     []string
	ChangedFields []TlDifferentField
	RemovedFields []string
}
