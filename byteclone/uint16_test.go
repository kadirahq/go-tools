package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeUint16(v uint16) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func TestUint16(t *testing.T) {
	v := NewUint16(nil)

	if !bytes.Equal(v.Bytes, BinaryEncodeUint16(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint16(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d := BinaryEncodeUint16(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:szuint16]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint16Read(b *testing.B) {
	var d = make([]byte, b.N*szuint16)
	var s = NewUint16(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*szuint16:])
	}
}

func BenchmarkUint16BinaryRead(b *testing.B) {
	var v uint16
	var d = make([]byte, b.N*szuint16)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint16Write(b *testing.B) {
	var s = NewUint16(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}

func BenchmarkUint16BinaryWrite(b *testing.B) {
	var v uint16
	var d = make([]byte, b.N*szuint16)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}
