package segfile

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
	"sync/atomic"

	goerr "github.com/go-errors/errors"
	"github.com/kadirahq/go-tools/fsutils"
	"github.com/kadirahq/go-tools/mmap"
	"github.com/kadirahq/go-tools/secure"
)

const (
	mdfile = "mdata"
)

var (
	// ErrSegSz is returned when segment file size is different
	ErrSegSz = errors.New("segment file size is different")

	// ErrOpts is returned when options have missing or invalid fields.
	ErrOpts = errors.New("invalid or missing options")

	// ErrMData is returned when metadata is invalid or corrupt
	ErrMData = errors.New("invalid or corrupt metadata")

	// ErrCorrupt is returned when segments are invalid or corrupt
	ErrCorrupt = errors.New("invalid or corrupt segments")

	// ErrParam is returned when given parameters are invalid
	ErrParam = errors.New("parameters are invalid")

	// ErrROnly is returned when attempt to write on read-only segfile
	ErrROnly = errors.New("segment file is read-only")

	// ErrClosed is returned when the resource is closed
	ErrClosed = errors.New("cannot use closed resource")
)

// Options for new File
type Options struct {
	// directory to store files
	// this field if required
	Path string

	// files will be prefixed
	Prefix string

	// size of a segment file
	FileSize int64

	// memory map segments
	MemoryMap bool

	// no writes allowed
	ReadOnly bool
}

// Defaults has values to use for missing fields
var defaults = &Options{
	Prefix:    "seg_",
	FileSize:  20 * 1024 * 1024,
	MemoryMap: false,
	ReadOnly:  false,
}

// Segment is a piece of the complete virtual file. A segment is stored
// in a file and can be loaded either directly or using memory maps.
type Segment interface {
	io.ReaderAt
	io.WriterAt
	io.Closer
	fsutils.Syncer
}

// File is similar to os.File but data is spread across many files.
// Data can be written and read directly or through memory mapping.
type File interface {
	io.Reader
	io.Writer
	io.ReaderAt
	io.WriterAt
	io.Seeker
	io.Closer
	fsutils.Syncer

	// Size returns pseudo segment file size.
	// This value will be usually less than the amount allocated on disk
	// This is used with io.Reader/io.Writer and for pre allocation.
	Size() (sz int64)

	// Grow grows the pseudo file size by sz bytes
	// New segment files will be allocated when necessary
	Grow(sz int64) (err error)

	// Reset sets offsets to 0
	Reset() (err error)

	// Clear sets file size to 0
	// Also set read-write offsets to zero.
	// This will not free up space on disk.
	Clear() (err error)
}

type file struct {
	// mutex to control allocating new segment files
	almutex sync.RWMutex

	// mutex to control io.Reader and io.Writer
	// TODO check whether this is necessary
	iomutex sync.RWMutex

	// metadata persister
	meta *Metadata

	// slice of segments (files or mmaps)
	segments []Segment

	// offset for io.Reader and io.Writer
	offset int64

	// flags which indicate whether we're using memory mapping
	// and whether this is a read only segmented file.
	mmap, ronly bool

	// struct states
	palloc *secure.Bool
	closed *secure.Bool

	// store necessary options for later use
	fpath   string
	fprefix string
	fsize   int64
	pthresh int64
}

// New creates a File struct with given options.
// Default values will be used for missing options.
func New(options *Options) (sf File, err error) {
	// validate options
	if options == nil ||
		options.Path == "" ||
		options.FileSize < 0 {
		return nil, goerr.Wrap(ErrOpts, 0)
	}

	// set default values for options
	if options.Prefix == "" {
		options.Prefix = defaults.Prefix
	}

	if options.FileSize == 0 {
		options.FileSize = defaults.FileSize
	}

	var meta *Metadata
	// path to metadata file
	mdpath := path.Join(options.Path, options.Prefix+mdfile)

	if !options.ReadOnly {
		if err := fsutils.EnsureDir(options.Path); err != nil {
			return nil, goerr.Wrap(err, 0)
		}
	}

	if options.ReadOnly {
		meta, err = ReadMetadata(mdpath)
	} else {
		meta, err = NewMetadata(mdpath, options.FileSize)
	}

	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	// validate metadata loaded from metadata file
	size, segs, used := meta.Size(), meta.Segs(), meta.Used()
	if size < 0 || segs < 0 || used < 0 || used > segs*size {
		return nil, goerr.Wrap(ErrMData, 0)
	}

	var segments []Segment
	bpath := path.Join(options.Path, options.Prefix)

	if options.MemoryMap {
		segments, err = loadMMaps(segs, size, bpath)
	} else {
		segments, err = loadFiles(segs, size, bpath)
	}

	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	// validate segments loaded from disk
	if loaded := len(segments); loaded != int(segs) {
		return nil, goerr.Wrap(ErrCorrupt, 0)
	}

	f := &file{
		meta:     meta,
		segments: segments,
		mmap:     options.MemoryMap,
		ronly:    options.ReadOnly,
		fpath:    options.Path,
		fprefix:  options.Prefix,
		fsize:    options.FileSize,
		pthresh:  options.FileSize / 2,
		palloc:   secure.NewBool(false),
		closed:   secure.NewBool(false),
	}

	// initial pre-allocation
	f.preallocate(f.pthresh)

	return f, nil
}

func (f *file) Read(p []byte) (n int, err error) {
	f.iomutex.Lock()
	defer f.iomutex.Unlock()

	n, err = f.ReadAt(p, f.offset)
	if err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	f.offset += int64(n)
	return n, nil
}

func (f *file) Write(p []byte) (n int, err error) {
	f.iomutex.Lock()
	defer f.iomutex.Unlock()

	n, err = f.WriteAt(p, f.offset)
	if err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	f.offset += int64(n)
	return n, nil
}

func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	meta := f.meta
	meta.RLock()
	defer meta.RUnlock()

	if f.closed.Get() {
		return 0, goerr.Wrap(ErrClosed, 0)
	}

	if p == nil || off < 0 {
		return 0, goerr.Wrap(ErrParam, 0)
	}

	size := len(p)
	if size == 0 {
		// empty read
		return 0, nil
	}

	sz64 := int64(size)
	segments := f.segments
	// get start/end segments and start/end offsets
	sseg, soff, eseg, eoff := calcRange(f.fsize, sz64, off)

	segs := meta.Segs()
	if sseg >= segs {
		return 0, io.EOF
	}

	// set data read to slice length by default
	// also handle EOF situations when reading
	n = size
	if eseg >= segs {
		eseg = segs - 1
		eoff = f.fsize

		segmentsRead := eseg - sseg
		fromFirstSeg := f.fsize - soff
		n = int(f.fsize*segmentsRead + fromFirstSeg)
	}

	for i := sseg; i <= eseg; i++ {
		var reader io.ReaderAt

		// offsets for reader
		var srcs, srce int64

		// offsets for result
		var dsts, dste int64

		// segment offset
		sego := i * f.fsize

		if i == sseg {
			srcs = soff
		} else {
			srcs = 0
		}

		if i == eseg {
			srce = eoff
		} else {
			srce = f.fsize
		}

		dsts = sego + srcs - off
		dste = sego + srce - off

		data := p[dsts:dste]
		reader = segments[i]

		n, err := reader.ReadAt(data, srcs)
		if err != nil {
			return 0, goerr.Wrap(err, 0)
		} else if n != len(data) {
			return 0, goerr.Wrap(fsutils.ErrReadSz, 0)
		}
	}

	return n, nil
}

func (f *file) WriteAt(p []byte, off int64) (n int, err error) {
	if f.closed.Get() {
		return 0, goerr.Wrap(ErrClosed, 0)
	}

	if f.ronly {
		return 0, goerr.Wrap(ErrROnly, 0)
	}

	if p == nil || off < 0 {
		return 0, goerr.Wrap(ErrParam, 0)
	}

	size := len(p)
	if size == 0 {
		// empty write
		return 0, nil
	}

	sz64 := int64(size)
	dend := off + sz64

	if err = f.ensureOffset(dend); err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	segments := f.segments
	// get start/end segments and start/end offsets
	sseg, soff, eseg, eoff := calcRange(f.fsize, sz64, off)

	for i := sseg; i <= eseg; i++ {
		var writer io.WriterAt

		// offsets for data
		var srcs, srce int64

		// offsets for writer
		var dsts, dste int64

		// segment offset
		sego := i * f.fsize

		if i == sseg {
			dsts = soff
		} else {
			dsts = 0
		}

		if i == eseg {
			dste = eoff
		} else {
			dste = f.fsize
		}

		srcs = sego + dsts - off
		srce = sego + dste - off

		data := p[srcs:srce]
		writer = segments[i]

		if n, err := writer.WriteAt(data, dsts); err != nil {
			return 0, goerr.Wrap(err, 0)
		} else if n != len(data) {
			return 0, goerr.Wrap(fsutils.ErrWriteSz, 0)
		}

		n = int(srce)
	}

	// update filesize
	f.updateSize(dend)

	// pre-allocate in a background go routine
	// pre allocation started only one at a time
	if f.palloc.Set(true) {
		go f.preallocate(dend + f.pthresh)
	}

	return n, nil
}

func (f *file) Seek(offset int64, whence int) (off int64, err error) {
	// TODO return correct error if offset < 0

	f.iomutex.Lock()
	defer f.iomutex.Unlock()

	meta := f.meta
	meta.RLock()
	defer meta.RUnlock()

	if f.closed.Get() {
		return 0, goerr.Wrap(ErrClosed, 0)
	}

	switch whence {
	case 0:
		off = offset
		atomic.StoreInt64(&f.offset, offset)
	case 1:
		off = atomic.AddInt64(&f.offset, offset)
	case 2:
		used := meta.Used()
		off = atomic.AddInt64(&f.offset, offset+used)
	}

	return off, nil
}

func (f *file) Size() (sz int64) {
	meta := f.meta
	meta.RLock()
	defer meta.RUnlock()

	if f.closed.Get() {
		return 0
	}

	used := meta.Used()
	return used
}

func (f *file) Grow(sz int64) (err error) {
	meta := f.meta

	if f.closed.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	// Calculate the file size after growing and make sure that the offset
	// exists in segfile. Allocate new segment files is necessary.
	// ensureOffset also locks the metadata file therefore, release the lock.
	meta.RLock()
	fsize := meta.Used() + sz
	meta.RUnlock()

	if err := f.ensureOffset(fsize); err != nil {
		return err
	}

	// update filesize
	f.updateSize(fsize)

	return nil
}

func (f *file) Reset() (err error) {
	f.iomutex.Lock()
	defer f.iomutex.Unlock()

	if f.closed.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	f.offset = 0
	return nil
}

func (f *file) Clear() (err error) {
	f.iomutex.Lock()
	defer f.iomutex.Unlock()

	meta := f.meta
	meta.Lock()
	defer meta.Unlock()

	if f.closed.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	// clear io.Reader/io.Writer offset
	f.offset = 0

	meta.SetUsed(0)
	meta.Sync()

	return nil
}

func (f *file) Sync() (err error) {
	if f.closed.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	for _, seg := range f.segments {
		err = seg.Sync()
		if err != nil {
			return goerr.Wrap(err, 0)
		}
	}

	return nil
}

func (f *file) Close() (err error) {
	f.almutex.Lock()
	defer f.almutex.Unlock()

	f.meta.Lock()
	defer f.meta.Unlock()

	if f.closed.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	// this will stop future requests
	f.closed.Set(true)

	if err = f.meta.Close(); err != nil {
		return goerr.Wrap(err, 0)
	}

	if f.mmap {
		closeMMaps(f.segments)
	} else {
		closeFiles(f.segments)
	}

	return nil
}

func (f *file) updateSize(off int64) {
	meta := f.meta
	meta.Lock()
	defer meta.Unlock()

	if f.closed.Get() {
		return
	}

	// used can change between RUnlock and Lock
	if used := meta.Used(); off > used {
		meta.SetUsed(off)
	}
}

func (f *file) preallocate(off int64) {
	defer f.palloc.Set(false)

	if f.closed.Get() {
		return
	}

	if err := f.ensureOffset(off); err != nil {
		// TODO handle pre allocation failure
	}
}

func (f *file) ensureOffset(off int64) (err error) {
	f.almutex.Lock()
	defer f.almutex.Unlock()

	meta := f.meta
	meta.Lock()
	defer meta.Unlock()

	if f.closed.Get() {
		return goerr.Wrap(ErrClosed, 0)
	}

	need := off / f.fsize

	if mod := off % f.fsize; mod != 0 {
		need++
	}

	have := int64(len(f.segments))
	diff := need - have

	if diff <= 0 {
		return nil
	}

	bpath := path.Join(f.fpath, f.fprefix)
	fsize := f.fsize

	segments := make([]Segment, need)
	copy(segments, f.segments)

	for i := have; i < need; i++ {
		var segment Segment
		fpath := bpath + strconv.Itoa(int(i))

		// make sure the file exist and has enough file size
		_, err = fsutils.EnsureFile(fpath, fsize)
		if err != nil {
			return goerr.Wrap(err, 0)
		}

		// load the segment
		if f.mmap {
			segment, err = loadMMap(fpath, fsize)
		} else {
			segment, err = loadFile(fpath, fsize)
		}

		if err != nil {
			return goerr.Wrap(err, 0)
		}

		segments[i] = segment
	}

	f.segments = segments
	meta.SetSegs(need)

	return nil
}

func calcRange(fsize, size, off int64) (sseg, soff, eseg, eoff int64) {
	sseg = off / fsize
	soff = off % fsize
	eseg = (size + off) / fsize
	eoff = (size + off) % fsize

	// if `eoff` is 0 there's no data to read from on `eseg`
	// `eseg` will be unavailable unless it's already allocated
	if eoff == 0 {
		eseg--
		eoff = fsize
	}

	return sseg, soff, eseg, eoff
}

// loadFiles loads a set of segment files (os.File) as segments
func loadFiles(segs, sz int64, bpath string) (segments []Segment, err error) {
	if segs <= 0 {
		segments = make([]Segment, 0)
		return segments, nil
	}

	segments = make([]Segment, segs)

	segsInt := int(segs)
	for i := 0; i < segsInt; i++ {
		fpath := bpath + strconv.Itoa(i)
		segment, err := loadFile(fpath, sz)
		if err != nil {
			closeFiles(segments)
			return nil, err
		}

		segments[i] = segment
	}

	return segments, nil
}

// loadFile loads a segment file at path and returns it
// It also ensures that these files are valid and has correct size.
func loadFile(fpath string, sz int64) (file *os.File, err error) {
	file, err = os.OpenFile(fpath, os.O_RDWR, 0644)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	finfo, err := file.Stat()
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	if finfo.IsDir() {
		return nil, goerr.Wrap(fsutils.ErrDirFile, 0)
	}

	if finfo.Size() != sz {
		return nil, goerr.Wrap(ErrSegSz, 0)
	}

	return file, nil
}

// closeFiles closes a slice of files
func closeFiles(files []Segment) {
	for _, file := range files {
		if file == nil {
			continue
		}

		if err := file.Close(); err != nil {
			// TODO handle file close fail
		}
	}
}

// loadMMaps loads a set of segment files (mmap.File) as segments
func loadMMaps(segs, sz int64, bpath string) (segments []Segment, err error) {
	if segs <= 0 {
		segments = make([]Segment, 0)
		return segments, nil
	}

	segments = make([]Segment, segs)

	segsInt := int(segs)
	for i := 0; i < segsInt; i++ {
		fpath := bpath + strconv.Itoa(i)
		segment, err := loadMMap(fpath, sz)
		if err != nil {
			closeMMaps(segments)
			return nil, err
		}

		segments[i] = segment
	}

	return segments, nil
}

// loadMMap loads a memory map of a segment file at path and returns it
// It also ensures that these mmaps are valid and has correct size.
func loadMMap(fpath string, sz int64) (mfile *mmap.File, err error) {
	mfile, err = mmap.NewFile(fpath, sz, true)
	if err != nil {
		return nil, goerr.Wrap(err, 0)
	}

	return mfile, nil
}

// closeMMaps closes a slice of mmaps
func closeMMaps(mmaps []Segment) {
	for _, mfile := range mmaps {
		if mfile == nil {
			continue
		}

		if err := mfile.Close(); err != nil {
			// TODO handle mmap close error
		}
	}
}
