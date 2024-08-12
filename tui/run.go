package tui

import (
	"TLExtractor/tui/types"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/kardianos/service"
	"slices"
)

func Run() {
	instance = &application{}
	instance.lg = lipgloss.DefaultRenderer()
	instance.styles = getStyle(instance.lg)
	var groups []*huh.Group
	var filteredMiniApps []*MiniApp
	var parentAppIds []int64
	for i, p := range miniApps {
		if p.parent != 0 {
			if !slices.Contains(parentAppIds, p.parent) {
				continue
			}
		} else {
			if err := p.check(types.InitCheck); err == nil {
				continue
			} else if !service.Interactive() {
				gologging.Fatal(err)
			}
		}
		if len(p.fields) == 0 {
			gologging.Fatal(fmt.Sprintf("Fields not set for %s", p.appName))
		}
		filteredMiniApps = append(filteredMiniApps, miniApps[i])
		parentAppIds = append(parentAppIds, p.id)
	}
	miniApps = filteredMiniApps
	for _, p := range miniApps {
		groups = append(groups, huh.NewGroup(p.fields...).WithHideFunc(p.hideFunc))
	}
	if len(groups) == 0 {
		return
	}
	instance.form = huh.NewForm(groups...).WithShowErrors(false).WithShowHelp(false)
	defaultKeyMap := huh.NewDefaultKeyMap()
	defaultKeyMap.FilePicker.Next = key.Binding{}
	defaultKeyMap.FilePicker.GoToTop.Keys()
	instance.form.WithKeyMap(defaultKeyMap)
	instance.programContext = tea.NewProgram(instance, tea.WithAltScreen())
	_, err := instance.programContext.Run()
	if err != nil {
		gologging.Fatal(err)
	}
}
