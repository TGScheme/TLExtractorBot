package tui

import (
	"reflect"
)

// TODO: This is a temporary fix for the page number retrieving
// Check the pr at https://github.com/charmbracelet/bubbletea/issues/1078
func (m *application) currentPage() int {
	return min(max(0, int(m.getSelector().MethodByName("Index").Call([]reflect.Value{})[0].Int())), len(miniApps)-1)
}
