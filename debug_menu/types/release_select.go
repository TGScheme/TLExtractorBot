package types

import (
	"TLExtractor/store_api/types"
	"fmt"
	"github.com/charmbracelet/huh"
)

type ReleaseSelect struct {
	*huh.Select[string]
	Value       string
	ReleaseList []types.Release
}

func (r *ReleaseSelect) NameFormat(release types.Release) string {
	return fmt.Sprintf("%s (%d)", release.Version, release.VersionCode/10)
}
