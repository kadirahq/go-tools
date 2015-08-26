package segfile

import (
	"io/ioutil"
	"sync"
	"time"

	goerr "github.com/go-errors/errors"
	fb "github.com/kadirahq/flatbuffers/go"
	"github.com/kadirahq/go-tools/fnutils"
	"github.com/kadirahq/go-tools/fsutils"
	"github.com/kadirahq/go-tools/logger"
	"github.com/kadirahq/go-tools/mmap"
	"github.com/kadirahq/go-tools/secure"
	"github.com/kadirahq/go-tools/segfile/metadata"
)

var (
	mdsize int64
	mdtemp []byte
)

func init() {
	// Create an empty metadata buffer which can be used as a template later.
	// When creating the table, always use non-zero values otherwise it will not
	// allocate space to store these fields. Set them to zero values later.

	b := fb.NewBuilder(0)
	metadata.MetadataStart(b)
	metadata.MetadataAddSegs(b, -1)
	metadata.MetadataAddSize(b, -1)
	metadata.MetadataAddUsed(b, -1)
	b.Finish(metadata.MetadataEnd(b))

	mdtemp = b.Bytes[b.Head():]
	mdsize = int64(len(mdtemp))

	meta := metadata.GetRootAsMetadata(mdtemp, 0)
	meta.SetSegs(0)
	meta.SetSize(0)
	meta.SetUsed(0)
}

// Metadata persists segfile information to disk in flatbuffer format
type Metadata struct {
	sync.RWMutex
	*metadata.Metadata

	memmap *mmap.File
	closed *secure.Bool
	syncfn *fnutils.Group
	dosync *secure.Bool
	rdonly bool
}

// NewMetadata creates a new metadata file at path
func NewMetadata(path string, sz int64) (m *Metadata, err error) {
	mfile, err := mmap.NewFile(path, 1, true)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	if mfile.Size() == 1 {
		if n, err := mfile.Write(mdtemp); err != nil {
			return nil, goerr.Wrap(err, 0)
		} else if n != len(mdtemp) {
			return nil, goerr.Wrap(fsutils.ErrWriteSz, 0)
		}
	}

	data := mfile.MMap.Data
	meta := metadata.GetRootAsMetadata(data, 0)
	if meta.Size() == 0 {
		meta.SetSize(sz)
	}

	batch := fnutils.NewGroup(func() {
		if err := mfile.Sync(); err != nil {
			logger.Error(err, "sync metadata")
		}
	})

	m = &Metadata{
		Metadata: meta,
		memmap:   mfile,
		closed:   secure.NewBool(false),
		dosync:   secure.NewBool(false),
		syncfn:   batch,
	}

	go func() {
		for _ = range time.Tick(10 * time.Millisecond) {
			if m.closed.Get() {
				break
			}

			if m.dosync.Get() {
				m.syncfn.Flush()
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	return m, nil
}

// ReadMetadata reads the file and parses metadata.
// Changes made to this metadata will not persist.
func ReadMetadata(path string) (mdata *Metadata, err error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	meta := metadata.GetRootAsMetadata(d, 0)
	mdata = &Metadata{
		Metadata: meta,
		closed:   secure.NewBool(false),
		rdonly:   true,
	}

	return mdata, nil
}

// Sync syncs the memory map to the disk
func (m *Metadata) Sync() {
	if !m.rdonly {
		m.dosync.Set(true)
		m.syncfn.Run()
	}
}

// Close closes metadata mmap file
func (m *Metadata) Close() (err error) {
	if m.closed.Get() {
		return nil
	}

	m.closed.Set(true)

	if !m.rdonly {
		err = m.memmap.Close()
		if err != nil {
			return goerr.Wrap(err, 0)
		}
	}

	return nil
}
