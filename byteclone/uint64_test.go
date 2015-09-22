package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeUint64(v uint64) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestUint64(t *testing.T) {
	v := NewUint64(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeUint64(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint64(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeUint64(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzUint64]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint64Read(b *testing.B) {
	var d = make([]byte, b.N*SzUint64)
	var s = NewUint64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzUint64:])
	}
}

func BenchmarkUint64BinaryRead(b *testing.B) {
	var v uint64
	var d = make([]byte, b.N*SzUint64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint64Write(b *testing.B) {
	var s = NewUint64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkUint64BinaryWrite(b *testing.B) {
	var v uint64
	var d = make([]byte, b.N*SzUint64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
