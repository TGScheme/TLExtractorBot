package types

type MultiPartInfo struct {
	Files map[string]FileDescriptor
	Data  map[string]string
}
