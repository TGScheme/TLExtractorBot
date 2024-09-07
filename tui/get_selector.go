package tui

import (
	"reflect"
	"unsafe"
)

func (m *application) getSelector() reflect.Value {
	v := reflect.ValueOf(m.form).Elem().FieldByName("selector").Elem()
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr()))
}
