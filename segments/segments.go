package segments

import (
	"errors"
	"io"

	"github.com/kadirahq/go-tools/fs"
)

var (
	// ErrSegSize is returned when the file size is different from the segment
	// size. In a segmented store, all segments should have the same file size.
	ErrSegSize = errors.New("wrong segment size")
)

// Store abstracts storing data in multiple segment files and provides it an
// interface which is similar to a single file. Benefits of using segmented
// files include faster writes, faster space allocation and smaller filesize.
// When slicing data, it will always stop when it hits a sgement boundary.
type Store interface {
	io.Reader
	io.Writer
	io.Seeker
	fs.Slicer
	io.ReaderAt
	io.WriterAt
	fs.SlicerAt
	fs.Syncer
	io.Closer
}

// BoundsFn is a function to execute for each segment.
// The loop will stop if this function returns an error.
// Runs with segment index and segment start/end offsets.
type BoundsFn func(i, start, end int64) (stop bool, err error)

// Bounds calculates segment boundaries and runs given function
// with starting and finishing offsets for each segment.
// This is not diomatic Go code but this loop code was reeated
// in many places in this package. Using a callback, we could
// reuse this set of code without doing any memory allocations.
func Bounds(size, start, end int64, fn BoundsFn) (err error) {
	ss := start / size
	so := start % size
	es := end / size
	eo := end % size

	if eo == 0 {
		eo = size
		es--
	}

	for i := ss; i <= es; i++ {
		var s, e int64

		// default
		e = size

		if i == ss {
			s = so
		}

		if i == es {
			e = eo
		}

		if stop, err := fn(i, s, e); err != nil {
			return err
		} else if stop {
			return nil
		}
	}

	return nil
}
