package segfile

import (
	"os"
	"testing"
)

var (
	MDPath = "/tmp/test-md"
)

func TestNewMetadata(t *testing.T) {
	if err := os.RemoveAll(MDPath); err != nil {
		t.Fatal(err)
	}

	md, err := NewMetadata(MDPath, 100)
	if err != nil {
		t.Fatal(err)
	}

	if got := md.Segs(); got != 0 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Size(); got != 100 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Used(); got != 0 {
		t.Fatal("incorrect values", got)
	}

	err = md.Close()
	if err != nil {
		t.Fatal(err)
	}

	md, err = NewMetadata(MDPath, 0)
	if err != nil {
		t.Fatal(err)
	}

	if got := md.Segs(); got != 0 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Size(); got != 100 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Used(); got != 0 {
		t.Fatal("incorrect values", got)
	}

	md.SetSegs(1)
	md.SetSize(2)
	md.SetUsed(3)

	if got := md.Segs(); got != 1 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Size(); got != 2 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Used(); got != 3 {
		t.Fatal("incorrect values", got)
	}

	err = md.Close()
	if err != nil {
		t.Fatal(err)
	}

	md, err = ReadMetadata(MDPath)
	if err != nil {
		t.Fatal(err)
	}

	if got := md.Segs(); got != 1 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Size(); got != 2 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Used(); got != 3 {
		t.Fatal("incorrect values", got)
	}

	md.SetSegs(100)
	md.SetSize(200)
	md.SetUsed(300)

	err = md.Close()
	if err != nil {
		t.Fatal(err)
	}

	md, err = ReadMetadata(MDPath)
	if err != nil {
		t.Fatal(err)
	}

	if got := md.Segs(); got != 1 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Size(); got != 2 {
		t.Fatal("incorrect values", got)
	}
	if got := md.Used(); got != 3 {
		t.Fatal("incorrect values", got)
	}

	err = md.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(MDPath); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkMetadataWrite(b *testing.B) {
	if err := os.RemoveAll(MDPath); err != nil {
		b.Fatal(err)
	}

	md, err := NewMetadata(MDPath, 100)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.SetParallelism(10000)
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for i = 0; pb.Next(); i++ {
			md.SetSegs(i)
		}
	})

	err = md.Close()
	if err != nil {
		b.Fatal(err)
	}

	if err := os.RemoveAll(MDPath); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkMetadataWriteAndSync(b *testing.B) {
	if err := os.RemoveAll(MDPath); err != nil {
		b.Fatal(err)
	}

	md, err := NewMetadata(MDPath, 100)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.SetParallelism(10000)
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for i = 0; pb.Next(); i++ {
			md.SetSegs(i)
			md.Sync()
		}
	})

	err = md.Close()
	if err != nil {
		b.Fatal(err)
	}

	if err := os.RemoveAll(MDPath); err != nil {
		b.Fatal(err)
	}
}
