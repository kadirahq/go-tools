package hybrid

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncodeUint32(v uint32) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func BinaryDecodeUint32(d []byte) uint32 {
	var v uint32
	b := bytes.NewBuffer(d)
	binary.Read(b, binary.LittleEndian, &v)
	return v
}

func TestUint32(t *testing.T) {
	var u uint32 = 5
	d := BinaryEncodeUint32(0)

	EncodeUint32(d, &u)
	if !bytes.Equal(d, BinaryEncodeUint32(5)) {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint32(10)
	DecodeUint32(d, &u)
	if u != BinaryDecodeUint32(d) {
		t.Fatal("wrong value")
	}

	// ---  ---  --- ---  ---  --- ---  ---  --- ---  ---  --- ---  ---

	v := NewUint32(nil)
	if !bytes.Equal(v.Bytes, BinaryEncodeUint32(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint32(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint32(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzUint32]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint32BinaryDecode(b *testing.B) {
	var v uint32
	var d = make([]byte, b.N*SzUint32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint32BinaryEncode(b *testing.B) {
	var v uint32
	var d = make([]byte, b.N*SzUint32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint32Decode(b *testing.B) {
	var d = make([]byte, SzUint32)
	var v uint32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeUint32(d, &v)
	}
}

func BenchmarkUint32Encode(b *testing.B) {
	var d = make([]byte, SzUint32)
	var v uint32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeUint32(d, &v)
	}
}

func BenchmarkUint32Read(b *testing.B) {
	var d = make([]byte, b.N*SzUint32)
	var s = NewUint32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzUint32:])
	}
}

func BenchmarkUint32Write(b *testing.B) {
	var s = NewUint32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}
