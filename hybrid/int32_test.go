package hybrid

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

func BinaryDecodeInt32(d []byte) int32 {
	var v int32
	b := bytes.NewBuffer(d)
	binary.Read(b, binary.LittleEndian, &v)
	return v
}

func TestInt32(t *testing.T) {
	var u int32 = 5
	d := BinaryEncodeInt32(0)

	EncodeInt32(d, &u)
	if !bytes.Equal(d, BinaryEncodeInt32(5)) {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeInt32(10)
	DecodeInt32(d, &u)
	if u != BinaryDecodeInt32(d) {
		t.Fatal("wrong value")
	}

	// ---  ---  --- ---  ---  --- ---  ---  --- ---  ---  --- ---  ---

	v := NewInt32(nil)
	if !bytes.Equal(v.Bytes, BinaryEncodeInt32(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeInt32(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeInt32(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzInt32]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkInt32BinaryDecode(b *testing.B) {
	var v int32
	var d = make([]byte, b.N*SzInt32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkInt32BinaryEncode(b *testing.B) {
	var v int32
	var d = make([]byte, b.N*SzInt32)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func BenchmarkInt32Decode(b *testing.B) {
	var d = make([]byte, SzInt32)
	var v int32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeInt32(d, &v)
	}
}

func BenchmarkInt32Encode(b *testing.B) {
	var d = make([]byte, SzInt32)
	var v int32

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeInt32(d, &v)
	}
}

func BenchmarkInt32Read(b *testing.B) {
	var d = make([]byte, b.N*SzInt32)
	var s = NewInt32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzInt32:])
	}
}

func BenchmarkInt32Write(b *testing.B) {
	var s = NewInt32(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}
