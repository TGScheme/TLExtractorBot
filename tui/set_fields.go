package tui

import "github.com/charmbracelet/huh"

func (miniApp *MiniApp) SetFields(fields ...huh.Field) {
	miniApp.fields = fields
}
