package types

import (
	"TLExtractor/appcenter/types"
	"fmt"
	"github.com/charmbracelet/huh"
	"strconv"
)

type ReleaseSelect struct {
	*huh.Select[string]
	Value       string
	ReleaseList []types.Release
}

func (r *ReleaseSelect) NameFormat(release types.Release) string {
	return fmt.Sprintf("%s (%s)", release.ShortVersion, release.Version[:4])
}

func (r *ReleaseSelect) GetReleaseId() string {
	for _, release := range r.ReleaseList {
		if r.Value == r.NameFormat(release) {
			return strconv.Itoa(release.Id)
		}
	}
	return ""
}
