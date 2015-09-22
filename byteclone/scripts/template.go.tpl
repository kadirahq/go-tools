package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	Sz{{BG}} = {{SZ}}
)

// Encode{{BG}} updates the byte slice to match value
func Encode{{BG}}(d []byte, v *{{SM}}) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*{{SM}})(unsafe.Pointer(head.Data))
	*value = *v
}

// Decode{{BG}} updates the value to match the byte slice
func Decode{{BG}}(d []byte, v *{{SM}}) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	value := (*{{SM}})(unsafe.Pointer(head.Data))
	*v = *value
}

// {{BG}} has a {{SM}} value and a byte slice using the same memory location.
// Any changes done to one of these fields will reflect on the other.
type {{BG}} struct {
	Value *{{SM}}
	Bytes []byte
}

// New{{BG}} will create a new {{BG}} struct with given byte slice.
// If the slice is nil, a new byte slice will be created for storage.
// If the slice length is less than required length, it will panic.
func New{{BG}}(d []byte) *{{BG}} {
	if d == nil {
		d = make([]byte, Sz{{BG}})
	}

	v := &{{BG}}{}
	v.Read(d[:Sz{{BG}}])
	return v
}

// Read updates the struct to use provided byte slice
// This can be used when it's required to read data from
func (v *{{BG}}) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*{{SM}})(unsafe.Pointer(head.Data))
	v.Bytes = d[:Sz{{BG}}]
}
