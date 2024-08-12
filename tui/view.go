package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func (m *application) View() string {
	s := m.styles
	logo := s.Logo.Render(miniApps[m.currentPage()].logo)
	var body strings.Builder
	if m.checking {
		centered := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height - 2).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center)
		body.WriteString(centered.Render(m.spinner.View()))
	} else {
		errors := m.form.Errors()
		if m.checkErr != nil {
			errors = append(errors, m.checkErr)
		}
		if m.checkFinalErr != nil {
			errors = append(errors, m.checkFinalErr)
		}
		var footerBuilder strings.Builder
		footerBuilder.WriteRune('\n')
		if len(errors) > 0 {
			for _, err := range errors {
				footerBuilder.WriteString(huh.ThemeCharm().Focused.ErrorMessage.Width(m.width - 2).Render(err.Error()))
			}
		} else {
			footerBuilder.WriteString(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		}
		footer := footerBuilder.String()
		m.form.WithHeight(m.height - lipgloss.Height(logo) - 2 - lipgloss.Height(footer))
		body.WriteString(logo)
		body.WriteRune('\n')
		body.WriteString(m.form.View())
		body.WriteString(footer)
	}

	return s.Base.
		Width(m.width).
		Height(m.height).
		Render(body.String())
}
