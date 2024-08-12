package tui

import (
	"github.com/charmbracelet/huh"
)

func (m *application) skipPage() {
	if m.checkErr != nil {
		page := m.currentPage()
		m.form.PrevGroup()
		if page > 0 {
			m.form.NextGroup()
		}
	} else {
		m.form.NextGroup()
	}
	if m.form.State == huh.StateCompleted {
		m.programContext.Quit()
	}
}
