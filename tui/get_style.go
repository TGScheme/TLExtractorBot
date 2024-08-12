package tui

import (
	"TLExtractor/tui/types"
	"github.com/charmbracelet/lipgloss"
)

var (
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
)

func getStyle(lg *lipgloss.Renderer) *types.Styles {
	s := types.Styles{}
	s.Base = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		Padding(1, 2)
	s.Logo = lg.NewStyle().
		Foreground(indigo).
		Margin(0, 0, 1, 0).
		Bold(true)
	return &s
}
