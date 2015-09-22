package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeInt32(v int32) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestInt32(t *testing.T) {
	v := NewInt32(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeInt32(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeInt32(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeInt32(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:szint32]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkInt32Read(b *testing.B) {
	var d = make([]byte, b.N*szint32)
	var s = NewInt32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*szint32:])
	}
}

func BenchmarkInt32BinaryRead(b *testing.B) {
	var v int32
	var d = make([]byte, b.N*szint32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkInt32Write(b *testing.B) {
	var s = NewInt32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkInt32BinaryWrite(b *testing.B) {
	var v int32
	var d = make([]byte, b.N*szint32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
