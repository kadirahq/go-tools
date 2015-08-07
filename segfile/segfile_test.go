package segfile

import (
	"os"
	"testing"
)

const (
	TmpDir = "/tmp/sfile"
)

func TestNew(t *testing.T) {
	err := os.RemoveAll(TmpDir)
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{Directory: TmpDir}
	sf, err := New(options)
	if err != nil {
		t.Fatal(err)
	}

	defer sf.Close()
}
