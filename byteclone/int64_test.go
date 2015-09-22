package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeInt64(v int64) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestInt64(t *testing.T) {
	v := NewInt64(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeInt64(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeInt64(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeInt64(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:szint64]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkInt64Read(b *testing.B) {
	var d = make([]byte, b.N*szint64)
	var s = NewInt64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*szint64:])
	}
}

func BenchmarkInt64BinaryRead(b *testing.B) {
	var v int64
	var d = make([]byte, b.N*szint64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkInt64Write(b *testing.B) {
	var s = NewInt64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkInt64BinaryWrite(b *testing.B) {
	var v int64
	var d = make([]byte, b.N*szint64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
