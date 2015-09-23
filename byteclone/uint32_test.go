package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeUint32(v uint32) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestUint32(t *testing.T) {
	v := NewUint32(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeUint32(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint32(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeUint32(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:szuint32]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint32Read(b *testing.B) {
	var d = make([]byte, b.N*szuint32)
	var s = NewUint32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*szuint32:])
	}
}

func BenchmarkUint32BinaryRead(b *testing.B) {
	var v uint32
	var d = make([]byte, b.N*szuint32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint32Write(b *testing.B) {
	var s = NewUint32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkUint32BinaryWrite(b *testing.B) {
	var v uint32
	var d = make([]byte, b.N*szuint32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
