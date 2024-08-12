package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *application) Init() tea.Cmd {
	return m.form.Init()
}
