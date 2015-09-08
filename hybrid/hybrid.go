package hybrid

import (
	"reflect"
	"unsafe"
)

type Uint32 struct {
	Value *uint32
	Bytes []byte
}

func NewUint32(v uint32) *Uint32 {
	data := make([]byte, 4)
	head := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	vptr := (*uint32)(unsafe.Pointer(head.Data))
	*vptr = v

	return &Uint32{vptr, data}
}

func ReadUint32(d []byte) *Uint32 {
	data := d[:4]
	head := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	vptr := (*uint32)(unsafe.Pointer(head.Data))

	return &Uint32{vptr, data}
}

func (s *Uint32) Read(d []byte) {
	head := (*reflect.SliceHeader)(unsafe.Pointer(&d))
	s.Value = (*uint32)(unsafe.Pointer(head.Data))
	s.Bytes = d[:4]
}
