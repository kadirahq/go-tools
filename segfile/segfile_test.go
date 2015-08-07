package segfile

import (
	"os"
	"runtime"
	"testing"
)

const (
	TmpDir = "/tmp/sfile"
)

func TNewOptions(t *testing.T, o *Options) {
	err := os.RemoveAll(TmpDir)
	if err != nil {
		t.Fatal(err)
	}

	sf, err := New(o)
	if err != nil {
		t.Fatal(err)
	}

	runtime.Gosched()

	err = sf.Close()
	if err != nil {
		t.Fatal(err)
	}

	sf, err = New(&Options{
		Directory:  o.Directory,
		FilePrefix: o.FilePrefix,
	})

	if err != nil {
		t.Fatal(err)
	}

	defer sf.Close()
}

func TestNewOptionsDefault(t *testing.T) {
	TNewOptions(t, &Options{Directory: TmpDir})
}

func TestNewOptionsAllOptions(t *testing.T) {
	TNewOptions(t, &Options{
		Directory:   TmpDir,
		FilePrefix:  "test_",
		SegmentSize: 5 * 1024 * 1024,
		MemoryMap:   true,
		ReadOnly:    false,
	})
}
