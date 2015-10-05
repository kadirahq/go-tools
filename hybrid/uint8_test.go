package hybrid

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

func BinaryDecodeUint8(d []byte) uint8 {
	var v uint8
	b := bytes.NewBuffer(d)
	binary.Read(b, binary.LittleEndian, &v)
	return v
}

func TestUint8(t *testing.T) {
	var u uint8 = 5
	d := BinaryEncodeUint8(0)

	EncodeUint8(d, &u)
	if !bytes.Equal(d, BinaryEncodeUint8(5)) {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint8(10)
	DecodeUint8(d, &u)
	if u != BinaryDecodeUint8(d) {
		t.Fatal("wrong value")
	}

	// ---  ---  --- ---  ---  --- ---  ---  --- ---  ---  --- ---  ---

	v := NewUint8(nil)
	if !bytes.Equal(v.Bytes, BinaryEncodeUint8(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint8(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint8(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzUint8]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint8BinaryDecode(b *testing.B) {
	var v uint8
	var d = make([]byte, b.N*SzUint8)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint8BinaryEncode(b *testing.B) {
	var v uint8
	var d = make([]byte, b.N*SzUint8)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint8Decode(b *testing.B) {
	var d = make([]byte, SzUint8)
	var v uint8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeUint8(d, &v)
	}
}

func BenchmarkUint8Encode(b *testing.B) {
	var d = make([]byte, SzUint8)
	var v uint8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeUint8(d, &v)
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

func BenchmarkUint8Write(b *testing.B) {
	var s = NewUint8(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}
