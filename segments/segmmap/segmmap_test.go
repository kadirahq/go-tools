package segmmap

import (
	"os"
	"testing"

	"github.com/kadirahq/go-tools/segments"
)

var (
	tmpdir  = "/tmp/test-segmmap/"
	tmpfile = tmpdir + "mmap_"
)

func setup(t *testing.T) {
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(tmpdir, 0777); err != nil {
		t.Fatal(err)
	}
}

func clear(t *testing.T) {
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	setup(t)
	defer clear(t)

	for i := 0; i < 3; i++ {
		s, err := New(tmpfile, 10)
		if err != nil {
			t.Fatal(err)
		}

		if len(s.segs) != 0 {
			t.Fatal("wrong length")
		}

		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestImpl(t *testing.T) {
	var s segments.Store
	s = &Store{}
	_ = s
}
