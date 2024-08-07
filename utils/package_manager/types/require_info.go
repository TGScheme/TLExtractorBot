package types

import "regexp"

type RequireInfo struct {
	Package     string
	File        string
	OnlyWindows bool
}

func (c RequireInfo) data() []string {
	return regexp.MustCompile(`([^/]+)`).FindAllString(c.Package, -1)
}

func (c RequireInfo) RepoOwner() string {
	return c.data()[0]
}

func (c RequireInfo) RepoName() string {
	return c.data()[1]
}

func (c RequireInfo) PackageName() string {
	if len(c.data()) > 2 {
		return c.data()[2]
	}
	return c.RepoName()
}
