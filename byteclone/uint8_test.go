package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeUint8(v uint8) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestUint8(t *testing.T) {
	v := NewUint8(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeUint8(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint8(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeUint8(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzUint8]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint8Read(b *testing.B) {
	var d = make([]byte, b.N*SzUint8)
	var s = NewUint8(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzUint8:])
	}
}

func BenchmarkUint8BinaryRead(b *testing.B) {
	var v uint8
	var d = make([]byte, b.N*SzUint8)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint8Write(b *testing.B) {
	var s = NewUint8(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkUint8BinaryWrite(b *testing.B) {
	var v uint8
	var d = make([]byte, b.N*SzUint8)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
