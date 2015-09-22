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

func BinaryDecodeInt64(d []byte) int64 {
	var v int64
	b := bytes.NewBuffer(d)
	binary.Read(b, binary.LittleEndian, &v)
	return v
}

func TestInt64(t *testing.T) {
	var u int64 = 5
	d := BinaryEncodeInt64(0)

	EncodeInt64(d, &u)
	if !bytes.Equal(d, BinaryEncodeInt64(5)) {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeInt64(10)
	DecodeInt64(d, &u)
	if u != BinaryDecodeInt64(d) {
		t.Fatal("wrong value")
	}

	// ---  ---  --- ---  ---  --- ---  ---  --- ---  ---  --- ---  ---

	v := NewInt64(nil)
	if !bytes.Equal(v.Bytes, BinaryEncodeInt64(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeInt64(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeInt64(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzInt64]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkInt64BinaryDecode(b *testing.B) {
	var v int64
	var d = make([]byte, b.N*SzInt64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkInt64BinaryEncode(b *testing.B) {
	var v int64
	var d = make([]byte, b.N*SzInt64)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func BenchmarkInt64Decode(b *testing.B) {
	var d = make([]byte, SzInt64)
	var v int64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeInt64(d, &v)
	}
}

func BenchmarkInt64Encode(b *testing.B) {
	var d = make([]byte, SzInt64)
	var v int64

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeInt64(d, &v)
	}
}

func BenchmarkInt64Read(b *testing.B) {
	var d = make([]byte, b.N*SzInt64)
	var s = NewInt64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzInt64:])
	}
}

func BenchmarkInt64Write(b *testing.B) {
	var s = NewInt64(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}
