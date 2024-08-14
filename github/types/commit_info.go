package types

type CommitInfo struct {
	SourceURL  string
	FilesLines map[string]string
}
