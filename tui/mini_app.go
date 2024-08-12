package tui

import (
	"TLExtractor/tui/types"
	"github.com/charmbracelet/huh"
)

type MiniApp struct {
	id             int64
	parent         int64
	appName        string
	logo           string
	fields         []huh.Field
	check          func(checkType types.CheckType) error
	hideFunc       func() bool
	loadingMessage string
}
