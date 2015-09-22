package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeFloat32(v float32) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestFloat32(t *testing.T) {
	v := NewFloat32(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeFloat32(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeFloat32(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeFloat32(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzFloat32]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkFloat32Read(b *testing.B) {
	var d = make([]byte, b.N*SzFloat32)
	var s = NewFloat32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzFloat32:])
	}
}

func BenchmarkFloat32BinaryRead(b *testing.B) {
	var v float32
	var d = make([]byte, b.N*SzFloat32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkFloat32Write(b *testing.B) {
	var s = NewFloat32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkFloat32BinaryWrite(b *testing.B) {
	var v float32
	var d = make([]byte, b.N*SzFloat32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
