package hybrid

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestUint32(t *testing.T) {
	s := NewUint32(1)

	if !bytes.Equal(s.Bytes, []byte{1, 0, 0, 0}) || *s.Value != 1 {
		t.Fatal("wrong value")
	}

	*s.Value = 2
	if !bytes.Equal(s.Bytes, []byte{2, 0, 0, 0}) || *s.Value != 2 {
		t.Fatal("wrong value")
	}

	*s.Value = 256
	if !bytes.Equal(s.Bytes, []byte{0, 1, 0, 0}) || *s.Value != 256 {
		t.Fatal("wrong value")
	}

	s = ReadUint32([]byte{1, 1, 0, 0, 0, 0, 0, 0})
	if !bytes.Equal(s.Bytes, []byte{1, 1, 0, 0}) || *s.Value != 257 {
		t.Fatal("wrong value")
	}

	s.Read([]byte{2, 1, 0, 0, 0, 0, 0, 0})
	if !bytes.Equal(s.Bytes, []byte{2, 1, 0, 0}) || *s.Value != 258 {
		t.Fatal("wrong value")
	}
}

func BenchmarkBinaryRead(b *testing.B) {
	var v uint32
	var d = make([]byte, b.N*4)
	var s = bytes.NewBuffer(d)

	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint32Read(b *testing.B) {
	var d = make([]byte, b.N*4)
	var s = NewUint32(0)

	for i := 0; i < b.N; i++ {
		s.Read(d[i*4:])
	}
}

func BenchmarkBinaryWrite(b *testing.B) {
	var v uint32
	var d = make([]byte, b.N*4)
	var s = bytes.NewBuffer(d)

	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint32Write(b *testing.B) {
	var s = NewUint32(0)
	for i := 0; i < b.N; i++ {
		*s.Value = uint32(i)
	}
}
