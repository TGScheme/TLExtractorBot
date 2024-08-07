package types

type TLObjDifferences struct {
	Object        TLInterface
	IsNew         bool
	NewFields     []string
	ChangedFields []TlDifferentField
	RemovedFields []string
}
