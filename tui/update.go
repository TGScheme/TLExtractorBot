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

type lazySkip struct{}
type cancelSkip struct{}
type lazyUpdate struct{}

func lazySkipPage() tea.Cmd {
	return func() tea.Msg {
		return lazySkip{}
	}
}

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
		go func(nextGroupMsg tea.Msg) {
			m.checkErr = page.check(types.SubmitCheck)
			if page.parent != 0 {
				parent := GetParentApp(page)
				if parent != nil {
					m.checkFinalErr = parent.check(types.FinalCheck)
				}
			}
			m.checking = false
			m.spinner = nil
			if m.checkErr != nil {
				m.pendingMsg = cancelSkip{}
			} else {
				m.pendingMsg = nextGroupMsg
			}
			if m.currentPage()+1 == len(miniApps) {
				m.programContext.Quit()
			}
			done <- true
		}(msg)
		select {
		case <-done:
		case <-timeout:
			m.checking = true
			m.spinner = spinner.New().Type(spinner.MiniDot)
			cmds = append(cmds, m.spinner.Init())
		}
		cmds = append(cmds, lazySkipPage())
		return m, tea.Batch(cmds...)
	}
	var cmds []tea.Cmd
	var wasLazySkip bool
	if _, ok := msg.(lazySkip); ok {
		if m.pendingMsg != nil {
			msg = m.pendingMsg
			wasLazySkip = true
			m.pendingMsg = nil
		} else {
			cmds = append(cmds, lazySkipPage())
		}
	}
	if m.spinner != nil {
		_, cmd := m.spinner.Title(lipgloss.NewStyle().PaddingLeft(1).Render(miniApps[m.currentPage()].loadingMessage)).Update(msg)
		cmds = append(cmds, cmd)
	}
	var wasLazyUpdate bool
	if _, ok := msg.(lazyUpdate); ok {
		wasLazyUpdate = true
		msg = tea.WindowSizeMsg{
			Width:  m.width + m.styles.Base.GetPaddingLeft(),
			Height: m.height + m.styles.Base.GetVerticalPadding(),
		}
	}
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
		if !wasLazyUpdate && wasLazySkip {
			cmds = append(cmds, func() tea.Msg {
				return lazyUpdate{}
			})
		}
	}
	return m, tea.Batch(cmds...)
}
