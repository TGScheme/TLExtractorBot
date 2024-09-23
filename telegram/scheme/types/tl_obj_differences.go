package types

type TLObjDifferences struct {
	Object        TLInterface
	IsNew         bool
	IsDeleted     bool
	NewFields     []string
	ChangedFields []TlDifferentField
	RemovedFields []string
}
