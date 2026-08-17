package types

type PatchInfo struct {
	OldConstructor     string `json:"old_constructor"`
	PatchedConstructor string `json:"new_constructor"`
}
