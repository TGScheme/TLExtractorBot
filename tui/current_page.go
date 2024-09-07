package tui

import (
	"reflect"
)

// TODO: This is a temporary fix for the page number retrieving
// Check the pr at https://github.com/charmbracelet/bubbletea/issues/1078
func (m *application) currentPage() int {
	reflectValue := reflect.ValueOf(*m.form)
	paginatorReflect := reflectValue.FieldByName("selector").Elem()
	return min(max(0, int(paginatorReflect.FieldByName("index").Int())), len(miniApps)-1)
}
