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

func BinaryDecodeUint64(d []byte) uint64 {
	var v uint64
	b := bytes.NewBuffer(d)
	binary.Read(b, binary.LittleEndian, &v)
	return v
}

func TestUint64(t *testing.T) {
	var u uint64 = 5
	d := BinaryEncodeUint64(0)

	EncodeUint64(d, &u)
	if !bytes.Equal(d, BinaryEncodeUint64(5)) {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint64(10)
	DecodeUint64(d, &u)
	if u != BinaryDecodeUint64(d) {
		t.Fatal("wrong value")
	}

	// ---  ---  --- ---  ---  --- ---  ---  --- ---  ---  --- ---  ---

	v := NewUint64(nil)
	if !bytes.Equal(v.Bytes, BinaryEncodeUint64(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint64(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint64(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzUint64]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint64BinaryDecode(b *testing.B) {
	var v uint64
	var d = make([]byte, b.N*SzUint64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint64BinaryEncode(b *testing.B) {
	var v uint64
	var d = make([]byte, b.N*SzUint64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint64Decode(b *testing.B) {
	var d = make([]byte, SzUint64)
	var v uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeUint64(d, &v)
	}
}

func BenchmarkUint64Encode(b *testing.B) {
	var d = make([]byte, SzUint64)
	var v uint64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeUint64(d, &v)
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

func BenchmarkUint64Write(b *testing.B) {
	var s = NewUint64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}
