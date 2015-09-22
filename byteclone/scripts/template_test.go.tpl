package byteclone

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BinaryEncode{{BG}}(v {{SM}}) []byte {
	b := bytes.NewBuffer(nil)
	binary.Write(b, binary.LittleEndian, v)
	return b.Bytes()
}

func BinaryDecode{{BG}}(d []byte) {{SM}} {
	var v {{SM}}
	b := bytes.NewBuffer(d)
	binary.Read(b, binary.LittleEndian, &v)
	return v
}

func Test{{BG}}(t *testing.T) {
	var u {{SM}} = 5
	d := BinaryEncode{{BG}}(0)

	Encode{{BG}}(d, &u)
	if !bytes.Equal(d, BinaryEncode{{BG}}(5)) {
		t.Fatal("wrong value")
	}

	d = BinaryEncode{{BG}}(10)
	Decode{{BG}}(d, &u)
	if u != BinaryDecode{{BG}}(d) {
		t.Fatal("wrong value")
	}

	// ---  ---  --- ---  ---  --- ---  ---  --- ---  ---  --- ---  ---

	v := New{{BG}}(nil)
	if !bytes.Equal(v.Bytes, BinaryEncode{{BG}}(0)) || *v.Value != 0 {
		t.Fatal("wrong value")
	}

	*v.Value = 5
	if !bytes.Equal(v.Bytes, BinaryEncode{{BG}}(5)) || *v.Value != 5 {
		t.Fatal("wrong value")
	}

	d = BinaryEncode{{BG}}(10)
	d = append(d, 1, 2, 3, 4, 5)
	v.Read(d)

	if !bytes.Equal(v.Bytes, d[:Sz{{BG}}]) || *v.Value != 10 {
		t.Fatal("wrong value")
	}
}

func Benchmark{{BG}}BinaryDecode(b *testing.B) {
	var v {{SM}}
	var d = make([]byte, b.N*Sz{{BG}})
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Read(s, binary.LittleEndian, v)
	}
}

func Benchmark{{BG}}BinaryEncode(b *testing.B) {
	var v {{SM}}
	var d = make([]byte, b.N*Sz{{BG}})
	var s = bytes.NewBuffer(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(s, binary.LittleEndian, v)
	}
}

func Benchmark{{BG}}Decode(b *testing.B) {
	var d = make([]byte, Sz{{BG}})
	var v {{SM}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode{{BG}}(d, &v)
	}
}

func Benchmark{{BG}}Encode(b *testing.B) {
	var d = make([]byte, Sz{{BG}})
	var v {{SM}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode{{BG}}(d, &v)
	}
}

func Benchmark{{BG}}Read(b *testing.B) {
	var d = make([]byte, b.N*Sz{{BG}})
	var s = New{{BG}}(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Read(d[i*Sz{{BG}}:])
	}
}

func Benchmark{{BG}}Write(b *testing.B) {
	var s = New{{BG}}(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		*s.Value = 0
	}
}
