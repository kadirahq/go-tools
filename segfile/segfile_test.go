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

	// run pre-alloc go routines
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

	err = sf.MemMap()
	if err != nil {
		t.Fatal(err)
	}

	err = sf.MemLock()
	if err != nil {
		t.Fatal(err)
	}

	err = sf.MUnlock()
	if err != nil {
		t.Fatal(err)
	}

	err = sf.MUnMap()
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	err = sf.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.RemoveAll(TmpDir)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewOptionsDefault(t *testing.T) {
	TNewOptions(t, &Options{Directory: TmpDir})
}

func TestNewOptionsFilePrefix(t *testing.T) {
	TNewOptions(t, &Options{Directory: TmpDir, FilePrefix: "test_"})
}

func TestNewOptionsReadOnly(t *testing.T) {
	TNewOptions(t, &Options{Directory: TmpDir, ReadOnly: true})
}

func TestNewOptionsSegSize(t *testing.T) {
	TNewOptions(t, &Options{Directory: TmpDir, SegmentSize: 5 * 1024 * 1024})
}

func TestNewOptionsMemMap(t *testing.T) {
	TNewOptions(t, &Options{Directory: TmpDir, MemoryMap: true})
}
