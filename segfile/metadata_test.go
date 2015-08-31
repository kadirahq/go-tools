package segfile

import (
	"os"
	"testing"

	"github.com/kadirahq/go-tools/logger"
)

var (
	MDPath = "/tmp/test-md"
)

func TestNewMetadata(t *testing.T) {
	if err := os.RemoveAll(MDPath); err != nil {
		logger.Error(err, "remove directory")
		t.Fatal(err)
	}

	md, err := NewMetadata(MDPath, 100)
	if err != nil {
		logger.Error(err, "create metadata")
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
		logger.Error(err, "close metadata")
		t.Fatal(err)
	}

	md, err = NewMetadata(MDPath, 0)
	if err != nil {
		logger.Error(err, "create metadata")
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

	md.MutateSegs(1)
	md.MutateSize(2)
	md.MutateUsed(3)

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
		logger.Error(err, "close metadata")
		t.Fatal(err)
	}

	md, err = ReadMetadata(MDPath)
	if err != nil {
		logger.Error(err, "read metadata")
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

	md.MutateSegs(100)
	md.MutateSize(200)
	md.MutateUsed(300)

	err = md.Close()
	if err != nil {
		logger.Error(err, "close metadata")
		t.Fatal(err)
	}

	md, err = ReadMetadata(MDPath)
	if err != nil {
		logger.Error(err, "read metadata")
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
		logger.Error(err, "close metadata")
		t.Fatal(err)
	}

	if err := os.RemoveAll(MDPath); err != nil {
		logger.Error(err, "remove directory")
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
	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for i = 0; pb.Next(); i++ {
			md.MutateSegs(i)
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
	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		var i int64
		for i = 0; pb.Next(); i++ {
			md.MutateSegs(i)
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
