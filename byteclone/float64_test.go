package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeFloat64(v float64) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestFloat64(t *testing.T) {
	v := NewFloat64(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeFloat64(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeFloat64(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeFloat64(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzFloat64]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkFloat64Read(b *testing.B) {
	var d = make([]byte, b.N*SzFloat64)
	var s = NewFloat64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzFloat64:])
	}
}

func BenchmarkFloat64BinaryRead(b *testing.B) {
	var v float64
	var d = make([]byte, b.N*SzFloat64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkFloat64Write(b *testing.B) {
	var s = NewFloat64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkFloat64BinaryWrite(b *testing.B) {
	var v float64
	var d = make([]byte, b.N*SzFloat64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
