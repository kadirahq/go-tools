package segfile

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/kadirahq/go-tools/logger"
	"github.com/kadirahq/go-tools/mdata"
	"github.com/kadirahq/go-tools/mmap"
)

const (
	// DirectoryPerm is the permission set for new directories
	DirectoryPerm = 0755

	// FileModeAlloc used when opening files for direct read/write
	FileModeAlloc = os.O_CREATE | os.O_RDWR

	// FileModeLoad used when opening files for direct read/write mode
	FileModeLoad = os.O_RDWR

	// FileModeRead used when opening files for direct read only mode
	FileModeRead = os.O_RDONLY

	// FilePerm used when opening files for direct read/write
	FilePerm = 0644

	// DefaultPrefix is used if the user does not provide a prefix
	// example: 'seg_0', 'seg_1', ... and 'seg_mdata' by default
	DefaultPrefix = "seg_"

	// MetadataFile is the filename used for metadata files
	// example: name will be 'seg_mdata' with default prefix
	MetadataFile = "mdata"

	// AllocThreshold is the percentage of a segment size used as a
	// threshold value to allocate a new segment file. When free space
	// goes below this threshold a new segment file will be allocated.
	AllocThreshold = 50

	// ChunkSize is the number of bytes to write at a time.
	// When creating new files, create it in small chunks.
	ChunkSize = 10 * 1024 * 1024
)

var (
	// ErrWrite is returned when bytes written is not equal to data size
	ErrWrite = errors.New("bytes written != data size")

	// ErrRead is returned when bytes read is not equal to data size
	ErrRead = errors.New("bytes read != data size")

	// ErrOptions is returned when options have missing or invalid fields.
	ErrOptions = errors.New("invalid or missing options")

	// ErrMData is returned when metadata is invalid or corrupt
	ErrMData = errors.New("invalid or corrupt metadata")

	// ErrSegDir is returned when segment file is a directory
	ErrSegDir = errors.New("segment file is a directory")

	// ErrSegSz is returned when segment file size is different
	ErrSegSz = errors.New("segment file size is different")

	// ErrClose is returned when close is closed multiple times
	ErrClose = errors.New("close called multiple times")

	// ErrClosed is returned when using closed segfile
	ErrClosed = errors.New("cannot use closed segfile")

	// ErrMapped is returned when files are already mapped
	ErrMapped = errors.New("files are already mapped")

	// ErrNotMapped is returned when files are not mapped
	ErrNotMapped = errors.New("files are not mapped")

	// ErrNoSeg is returned when segment is missing
	ErrNoSeg = errors.New("segment is missing")

	// ErrROnly is returned when attempt to write on read-only segfile
	ErrROnly = errors.New("segment file is read-only")

	// ErrParams is returned when given parameters are invalid
	ErrParams = errors.New("parameters are invalid")

	// ChunkBytes is a ChunkSize size slice of zeroes
	ChunkBytes = make([]byte, ChunkSize)

	// Logger logs stuff
	Logger = logger.New("SEGFILE")
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

// DefaultOptions has values to use for missing fields
var DefaultOptions = &Options{
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

	// Close closes the segment
	Close() (err error)
}

// File is similar to os.File but data is spread across many files.
// Data can be written and read directly or through memory mapping.
type File interface {
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt

	// Info returns segment file metadata
	Info() (meta *Metadata)

	// Grow grows the pseudo file size by sz bytes
	// New segment files are automatically allocated when
	// available space is not enough.
	Grow(sz int64) (err error)

	// Reset sets offsets to 0
	Reset() (err error)

	// Clear sets file size to 0
	// Also set read-write offsets to zero.
	// This will not free up space on disk.
	Clear() (err error)

	// Close cleans up everything and closes files
	Close() (err error)
}

type file struct {
	// metadata struct
	meta *Metadata

	// metadata helper
	mdata mdata.Data

	// slice of segments (files or mmaps)
	segments []Segment

	// allocation mutex to control allocations
	almutex *sync.Mutex

	// set to true when the segments are memory mapped
	mmapped bool

	// set to true when the file is closed
	closed bool

	// set to true when the file is read only
	ronly bool

	// set to true while a pre allocation is running
	prealloc bool

	// io.Reader/io.Writer offset
	rwoffset int64
}

// New creates a File struct with given options.
// Default values will be used for missing options.
func New(options *Options) (sf File, err error) {
	// validate options
	if options == nil ||
		options.Path == "" ||
		options.FileSize < 0 {
		Logger.Trace(ErrOptions)
		return nil, ErrOptions
	}

	// set default values for options
	if options.Prefix == "" {
		options.Prefix = DefaultOptions.Prefix
	}

	if options.FileSize == 0 {
		options.FileSize = DefaultOptions.FileSize
	}

	// make sure target directory exists
	err = os.MkdirAll(options.Path, DirectoryPerm)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	// create/load metadata
	meta := &Metadata{
		Path:     options.Path,
		Prefix:   options.Prefix,
		FileSize: options.FileSize,
	}

	mdpath := path.Join(options.Path, options.Prefix+MetadataFile)
	md, err := mdata.New(mdpath, meta, options.ReadOnly)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	// validate metadata file
	if meta.Path != options.Path ||
		meta.Prefix != options.Prefix ||
		meta.FileSize < 0 ||
		meta.Segments < 0 ||
		meta.DataSize < 0 ||
		meta.DataSize > meta.Segments*meta.FileSize {
		Logger.Trace(ErrMData)
		return nil, ErrMData
	}

	f := &file{
		meta:    meta,
		mdata:   md,
		almutex: &sync.Mutex{},
		ronly:   options.ReadOnly,
	}

	if options.MemoryMap {
		f.mmapped = true
		err = f.loadMMaps()
	} else {
		err = f.loadFiles()
	}

	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	if !f.ronly {
		f.preallocateIfNeeded()
	}

	return f, nil
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	n, err = f.ReadAt(p, f.rwoffset)
	if err != nil {
		Logger.Trace(err)
		return 0, err
	}

	atomic.AddInt64(&f.rwoffset, int64(n))
	return n, nil
}

func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	if p == nil || off < 0 {
		Logger.Trace(ErrParams)
		return 0, ErrParams
	}

	if len(p) == 0 {
		// empty read
		return 0, nil
	}

	meta := f.meta
	size := len(p)
	sz64 := int64(size)
	sseg := off / meta.FileSize
	soff := off % meta.FileSize
	eseg := (sz64 + off) / meta.FileSize
	eoff := (sz64 + off) % meta.FileSize

	// if `eoff` is 0 there's no data to read from on `eseg`
	// `eseg` will be unavailable unless it's already allocated
	if eoff == 0 {
		eseg--
		eoff = meta.FileSize
	}

	if sseg >= meta.Segments {
		return 0, io.EOF
	}

	if eseg < meta.Segments {
		n = size
	} else {
		eseg = meta.Segments
		eoff = meta.FileSize
		n = int(meta.FileSize*(eseg-sseg) + meta.FileSize - soff)
	}

	for i := sseg; i <= eseg; i++ {
		var reader io.ReaderAt
		var srcStart, srcEnd int64

		if i == sseg {
			srcStart = soff
		} else {
			srcStart = 0
		}

		if i == eseg {
			srcEnd = eoff
		} else {
			srcEnd = meta.FileSize
		}

		segStart := i * meta.FileSize
		dstStart := segStart + srcStart - off
		dstEnd := segStart + srcEnd - off
		data := p[dstStart:dstEnd]
		reader = f.segments[i]

		n, err := reader.ReadAt(data, srcStart)
		if err != nil {
			Logger.Trace(err)
			return 0, err
		} else if n != len(data) {
			Logger.Trace(ErrRead)
			return 0, ErrRead
		}
	}

	return n, nil
}

func (f *file) Write(p []byte) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	n, err = f.WriteAt(p, f.rwoffset)
	if err != nil {
		Logger.Trace(err)
		return 0, err
	}

	atomic.AddInt64(&f.rwoffset, int64(n))
	return n, nil
}

func (f *file) WriteAt(p []byte, off int64) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	if f.ronly {
		Logger.Trace(ErrROnly)
		return 0, ErrROnly
	}

	if p == nil || off < 0 {
		Logger.Trace(ErrParams)
		return 0, ErrParams
	}

	if len(p) == 0 {
		// empty write
		return 0, nil
	}

	// pre-allocate in background go routine
	// go routine started only if necessary
	f.preallocateIfNeeded()

	meta := f.meta
	size := int64(len(p))

	// additional space required for write
	// allocated in current go routine (before write)
	total := meta.Segments * meta.FileSize
	if sz := off + size - total; sz > 0 {
		err = f.ensureSpace(sz)
		if err != nil {
			Logger.Trace(err)
			return 0, err
		}
	}

	sseg := off / meta.FileSize
	soff := off % meta.FileSize
	eseg := (size + off) / meta.FileSize
	eoff := (size + off) % meta.FileSize

	// if `eoff` is 0 there's no data to read from on `eseg`
	// `eseg` will be unavailable unless it's already allocated
	if eoff == 0 {
		eseg--
		eoff = meta.FileSize
	}

	for i := sseg; i <= eseg; i++ {
		var writer io.WriterAt
		var dstStart, dstEnd int64

		if i == sseg {
			dstStart = soff
		} else {
			dstStart = 0
		}

		if i == eseg {
			dstEnd = eoff
		} else {
			dstEnd = meta.FileSize
		}

		segStart := i * meta.FileSize
		srcStart := segStart + dstStart - off
		srcEnd := segStart + dstEnd - off
		data := p[srcStart:srcEnd]
		writer = f.segments[i]

		num, err := writer.WriteAt(data, dstStart)
		if err != nil {
			Logger.Trace(err)
			return 0, err
		} else if num != len(data) {
			Logger.Trace(ErrWrite)
			return 0, ErrWrite
		}

		n = int(srcEnd)
	}

	return n, nil
}

func (f *file) Info() (meta *Metadata) {
	return f.meta
}

func (f *file) Grow(sz int64) (err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return ErrClosed
	}

	err = f.ensureSpace(sz)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	f.meta.DataSize += sz

	err = f.mdata.Save()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	return nil
}

func (f *file) Reset() (err error) {
	if f.closed {
		Logger.Error(ErrClose)
		return ErrClose
	}

	atomic.StoreInt64(&f.rwoffset, 0)
	return nil
}

func (f *file) Clear() (err error) {
	if f.closed {
		Logger.Trace(ErrClose)
		return ErrClose
	}

	atomic.StoreInt64(&f.rwoffset, 0)
	atomic.StoreInt64(&f.meta.DataSize, 0)

	err = f.mdata.Save()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	return nil
}

func (f *file) Close() (err error) {
	if f.closed {
		Logger.Error(ErrClose)
		return nil
	}

	f.almutex.Lock()
	defer f.almutex.Unlock()

	f.closed = true

	err = f.mdata.Close()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	if f.mmapped {
		closeMMaps(f.segments)
	} else {
		closeFiles(f.segments)
	}

	return nil
}

// shouldAllocate checks whether there's free space to store sz bytes
func (f *file) shouldAllocate(sz int64) (do bool) {
	if f.closed {
		return false
	}

	meta := f.meta
	total := meta.FileSize * meta.Segments
	return meta.DataSize+sz > total
}

// preallocateIfNeeded pre allocates new segment files if free space
// goes below the threshold (AllocThreshold percentage of FileSize).
func (f *file) preallocateIfNeeded() {
	meta := f.meta
	thresh := meta.FileSize * AllocThreshold / 100

	// TODO Ensure only one preallocate go routine is run.
	//      It is possible multiple go routines to pass the
	//      if !f.prealloc test before setting it to true.
	//      This is a rare case and will not cause problems.
	//      Find a way to make sure only one gets through.
	//      This method is called with all write requests
	//      the method used should be faster than accidentally
	//      starting an additional very lightweight go routine.

	// Return if a pre-allocation is already in progress.
	if !f.prealloc && f.shouldAllocate(thresh) {
		// set allocing to true before starting pre allocation goroutine
		// starting many unnecessary go routines can be extremely costly
		f.prealloc = true

		go func() {
			if err := f.ensureSpace(thresh); err != nil {
				Logger.Error(err)
			}

			f.prealloc = false
		}()
	}
}

// ensureSpace makes sure there's enough space left to store an
// additional sz bytes. New segment files will be created if needed.
func (f *file) ensureSpace(sz int64) (err error) {
	// use the check-lock-check method to reduce locks
	// checking is cheaper than always locking the mutex
	if f.shouldAllocate(sz) {
		f.almutex.Lock()
		defer f.almutex.Unlock()

		if f.shouldAllocate(sz) {
			if err := f.allocateSpace(sz); err != nil {
				Logger.Trace(err)
				return err
			}
		}
	}

	return nil
}

// allocateSpace allocates additional space at segfile end
// One or more segment files will be created to hold sz bytes
func (f *file) allocateSpace(sz int64) (err error) {
	meta := f.meta
	pathPrefix := path.Join(meta.Path, meta.Prefix)
	filesCount := sz / meta.FileSize

	if sz%meta.FileSize != 0 {
		filesCount++
	}

	var i int64
	for i = 0; i < filesCount; i++ {
		segID := meta.Segments + i
		idStr := strconv.Itoa(int(segID))
		fpath := pathPrefix + idStr

		err = createFile(fpath, meta.FileSize)
		if err != nil {
			Logger.Trace(err)
			return err
		}

		var segment Segment

		if f.mmapped {
			segment, err = loadMMap(fpath, meta.FileSize)
		} else {
			segment, err = loadFile(fpath, meta.FileSize)
		}

		if err != nil {
			Logger.Trace(err)
			return err
		}

		f.segments = append(f.segments, segment)
	}

	meta.Segments += filesCount

	err = f.mdata.Save()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	return nil
}

// loadFiles loads all segment files according to segfile metadata.
// It also ensures all segment files are valid.
func (f *file) loadFiles() (err error) {
	meta := f.meta
	f.segments = make([]Segment, meta.Segments)

	if meta.Segments > 0 {
		var i int64
		for i = 0; i < meta.Segments; i++ {
			istr := strconv.Itoa(int(i))
			fpath := path.Join(meta.Path, meta.Prefix+istr)

			segment, err := loadFile(fpath, meta.FileSize)
			if err != nil {
				Logger.Trace(err)
				closeFiles(f.segments)
				return err
			}

			f.segments[i] = segment
		}
	}

	return nil
}

// loadMMaps loads all segment files as memory maps according to
// segfile metadata. It also ensures all memory maps are valid.
func (f *file) loadMMaps() (err error) {
	meta := f.meta
	f.segments = make([]Segment, meta.Segments)

	if meta.Segments > 0 {
		var i int64
		for i = 0; i < meta.Segments; i++ {
			istr := strconv.Itoa(int(i))
			fpath := path.Join(meta.Path, meta.Prefix+istr)

			segment, err := loadMMap(fpath, meta.FileSize)
			if err != nil {
				Logger.Trace(err)
				closeMMaps(f.segments)
				return err
			}

			f.segments[i] = segment
		}
	}

	return nil
}

// loadFile loads a segment file at path and returns it
// It also ensures that these files are valid and has correct size.
func loadFile(fpath string, sz int64) (file *os.File, err error) {
	file, err = os.OpenFile(fpath, FileModeLoad, FilePerm)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	finfo, err := file.Stat()
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	if finfo.IsDir() {
		Logger.Trace(ErrSegDir)
		return nil, ErrSegDir
	}

	if finfo.Size() != sz {
		Logger.Trace(ErrSegSz)
		return nil, ErrSegSz
	}

	return file, nil
}

// loadMMap loads a memory map of a segment file at path and returns it
// It also ensures that these mmaps are valid and has correct size.
func loadMMap(fpath string, sz int64) (mfile mmap.File, err error) {
	mopts := &mmap.Options{
		Path: fpath,
		Size: sz,
	}

	mfile, err = mmap.New(mopts)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	err = mfile.Lock()
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	return mfile, nil
}

// closeFiles closes a slice of files
func closeFiles(files []Segment) {
	if files == nil {
		return
	}

	for _, file := range files {
		if file == nil {
			continue
		}

		err := file.Close()
		if err != nil {
			Logger.Error(err)
		}
	}
}

// closeMMaps closes a slice of mmaps
func closeMMaps(mmaps []Segment) {
	if mmaps == nil {
		return
	}

	for _, mfile := range mmaps {
		if mfile == nil {
			continue
		}

		err := mfile.Close()
		if err != nil {
			Logger.Error(err)
		}
	}
}

// createFile creates a new file at `fpath` with size `sz`
// After creating the file, it's content should be all zeroes.
func createFile(fpath string, sz int64) (err error) {
	file, err := os.OpenFile(fpath, FileModeAlloc, FilePerm)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	// number of chunks to write
	chunks := sz / ChunkSize

	var i int64
	for i = 0; i < chunks; i++ {
		offset := ChunkSize * i
		n, err := file.WriteAt(ChunkBytes, offset)
		if err != nil {
			Logger.Trace(err)
			return err
		} else if int64(n) != ChunkSize {
			Logger.Trace(ErrWrite)
			return ErrWrite
		}
	}

	// write all remaining bytes
	toWrite := sz % ChunkSize
	zeroes := ChunkBytes[:toWrite]
	offset := ChunkSize * chunks
	n, err := file.WriteAt(zeroes, offset)
	if err != nil {
		Logger.Trace(err)
		return err
	} else if int64(n) != toWrite {
		Logger.Trace(ErrWrite)
		return ErrWrite
	}

	return nil
}
