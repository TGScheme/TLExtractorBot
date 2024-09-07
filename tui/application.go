package tui

import (
	"TLExtractor/tui/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

var instance *application
var miniApps []*MiniApp

type application struct {
	width, height           int
	programContext          *tea.Program
	lg                      *lipgloss.Renderer
	styles                  *types.Styles
	form                    *huh.Form
	spinner                 *spinner.Spinner
	pendingMsg              tea.Msg
	isBack                  bool
	checking                bool
	checkErr, checkFinalErr error
}
