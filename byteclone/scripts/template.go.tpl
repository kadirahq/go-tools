package byteclone

import (
	"reflect"
	"unsafe"
)

const (
	sz{{SM}} = {{SZ}}
)

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
		d = make([]byte, sz{{SM}})
	}

	v := &{{BG}}{}
	v.Read(d[:sz{{SM}}])
	return v
}

func (v *{{BG}}) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	v.Value = (*{{SM}})(unsafe.Pointer(head.Data))
	v.Bytes = d[:sz{{SM}}]
}
