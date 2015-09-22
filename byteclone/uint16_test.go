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

func BinaryDecodeUint16(d []byte) uint16 {
	var v uint16
	b := bytes.NewBuffer(d)
	binary.Read(b, binary.LittleEndian, &v)
	return v
}

func TestUint16(t *testing.T) {
	var u uint16 = 5
	d := BinaryEncodeUint16(0)

	EncodeUint16(d, &u)
	if !bytes.Equal(d, BinaryEncodeUint16(5)) {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint16(10)
	DecodeUint16(d, &u)
	if u != BinaryDecodeUint16(d) {
		t.Fatal("wrong value")
	}

	// ---  ---  --- ---  ---  --- ---  ---  --- ---  ---  --- ---  ---

	v := NewUint16(nil)
	if !bytes.Equal(v.Bytes, BinaryEncodeUint16(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncodeUint16(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d = BinaryEncodeUint16(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:SzUint16]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func BenchmarkUint16BinaryDecode(b *testing.B) {
	var v uint16
	var d = make([]byte, b.N*SzUint16)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint16BinaryEncode(b *testing.B) {
	var v uint16
	var d = make([]byte, b.N*SzUint16)
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func BenchmarkUint16Decode(b *testing.B) {
	var d = make([]byte, SzUint16)
	var v uint16

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeUint16(d, &v)
	}
}

func BenchmarkUint16Encode(b *testing.B) {
	var d = make([]byte, SzUint16)
	var v uint16

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeUint16(d, &v)
	}
}

func BenchmarkUint16Read(b *testing.B) {
	var d = make([]byte, b.N*SzUint16)
	var s = NewUint16(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*SzUint16:])
	}
}

func BenchmarkUint16Write(b *testing.B) {
	var s = NewUint16(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}
