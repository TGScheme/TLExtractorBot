package tui

import (
	"TLExtractor/tui/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"reflect"
	"time"
)

func (m *application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width - m.styles.Base.GetPaddingLeft()
		m.height = msg.Height - m.styles.Base.GetVerticalPadding()
	case tea.KeyMsg:
		m.isBack = false
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			return m, tea.Quit
		case "shift+tab":
			m.isBack = true
		}
		m.checkErr = nil
	}
	switch reflect.TypeOf(msg).String() {
	case "huh.nextGroupMsg":
		var cmds []tea.Cmd
		done := make(chan bool)
		page := miniApps[m.currentPage()]
		timeout := time.After(50 * time.Millisecond)
		go func() {
			m.checkErr = page.check(types.SubmitCheck)
			if page.parent != 0 {
				parent := GetParentApp(page)
				if parent != nil {
					m.checkFinalErr = parent.check(types.FinalCheck)
				}
			}
			m.checking = false
			m.spinner = nil
			m.skipPage()
			done <- true
		}()
		select {
		case <-done:
		case <-timeout:
			m.checking = true
			m.spinner = spinner.New().
				Title(lipgloss.NewStyle().PaddingLeft(1).Render(page.loadingMessage)).
				Type(spinner.MiniDot)
			cmds = append(cmds, m.spinner.Init())
		}
		return m, tea.Batch(cmds...)
	}
	var cmds []tea.Cmd
	if m.spinner != nil {
		_, cmd := m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}
