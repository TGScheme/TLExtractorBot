package types

type PackageInfo struct {
	Name        string
	Owner       string
	DownloadURL string
	Version     string
	Size        int
	FileName    string
	Path        string
	Variant     string
}

func (p PackageInfo) GetFullName() string {
	return p.Name + "-" + p.Version
}
