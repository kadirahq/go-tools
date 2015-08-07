package segfile

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/kadirahq/go-tools/logger"
	"github.com/kadirahq/go-tools/mdata"
	"github.com/kadirahq/go-tools/mmap"
)

const (
	// DirectoryPerm is the permission set for new directories
	DirectoryPerm = 0755

	// FileModeAlloc used when opening files for direct read/write
	FileModeAlloc = os.O_CREATE | os.O_RDWR

	// FileModeLoad used when opening files for direct read/write
	FileModeLoad = os.O_RDWR

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
	AllocThreshold = 10

	// ChunkSize is the number of bytes to write at a time.
	// When creating new files, create it in small chunks.
	ChunkSize = 10 * 1024 * 1024
)

var (
	// ErrWrite is returned when bytes written not equal to data size
	ErrWrite = errors.New("bytes written != data size")

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

	// ChunkBytes is a ChunkSize size slice of zeroes
	ChunkBytes = make([]byte, ChunkSize)

	// Logger logs stuff
	Logger = logger.New("SEGFILE")
)

// Options for new File
type Options struct {
	// directory to store files
	// this field if required
	Directory string

	// files will be prefixed
	FilePrefix string

	// size of a segment file
	SegmentSize int64

	// memory map by default
	MemoryMap bool

	// no writes allowed
	ReadOnly bool
}

// DefaultOptions has values to use for missing fields
var DefaultOptions = &Options{
	FilePrefix:  "seg_",
	SegmentSize: 100 * 1024 * 1024,
	MemoryMap:   false,
	ReadOnly:    false,
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

	// MemMap maps all files and switches to mmap mode
	// All reads and writes will be performed with mmaps
	MemMap(lock bool) (err error)

	// MUnmap unmaps all maps and switches to file mode
	// All reads and writes will be performed with files
	MUnmap() (err error)

	// Close cleans up everything and closes files
	Close() (err error)
}

type file struct {
	// metadata struct
	meta *Metadata

	// metadata helper
	mdata mdata.Data

	// set to true when memory maps are used
	mapped bool

	// slice of files used for direct read/write
	files []*os.File

	// slice of memory maps for mapped read/write
	mmaps []mmap.File

	// read/write mutex
	rwmutex *sync.RWMutex

	// growth mutex to control Grow calls
	grmutex *sync.Mutex

	// allocation mutex to control allocations
	almutex *sync.Mutex

	// set to true when the file is closed
	closed bool

	// set to true while a pre allocation is running
	prealloc bool

	// io.Reader read offset
	roffset int64

	// io.Reader write offset
	woffset int64
}

// New creates a File struct with given options.
// Default values will be used for missing options.
func New(options *Options) (sf File, err error) {
	// validate options
	if options == nil ||
		options.Directory == "" ||
		options.SegmentSize < 0 {
		Logger.Trace(ErrOptions)
		return nil, ErrOptions
	}

	// set default values for options
	if options.FilePrefix == "" {
		options.FilePrefix = DefaultOptions.FilePrefix
	}

	if options.SegmentSize == 0 {
		options.SegmentSize = DefaultOptions.SegmentSize
	}

	// make sure target directory exists
	err = os.MkdirAll(options.Directory, DirectoryPerm)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	// create/load metadata
	meta := &Metadata{
		Directory:   options.Directory,
		FilePrefix:  options.FilePrefix,
		SegmentSize: options.SegmentSize,
	}

	mdpath := path.Join(options.Directory, options.FilePrefix+MetadataFile)
	md, err := mdata.New(mdpath, meta, options.ReadOnly)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	// validate metadata file
	if meta.Directory != options.Directory ||
		meta.FilePrefix != options.FilePrefix ||
		meta.SegmentSize < 0 ||
		meta.SegmentFiles < 0 ||
		meta.DataSize < 0 ||
		meta.DataSize > meta.SegmentFiles*meta.SegmentSize {
		Logger.Trace(ErrMData)
		return nil, ErrMData
	}

	f := &file{
		meta:    meta,
		mdata:   md,
		rwmutex: &sync.RWMutex{},
		almutex: &sync.Mutex{},
		grmutex: &sync.Mutex{},
	}

	if options.MemoryMap {
		f.mapped = true
		err = f.loadMMaps()
	} else {
		err = f.loadFiles()
	}

	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	if !options.ReadOnly {
		// initial pre-allocation
		go f.preallocateIfNeeded()
	}

	return f, nil
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	f.rwmutex.RLock()
	defer f.rwmutex.RUnlock()

	n, err = f.ReadAt(p, f.roffset)
	if err != nil {
		Logger.Trace(err)
		return 0, err
	}

	f.roffset += int64(n)
	return 0, nil
}

func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	f.rwmutex.RLock()
	defer f.rwmutex.RUnlock()

	if f.mapped {
		n, err = f.readMMaps(p, off)
	} else {
		n, err = f.readFiles(p, off)
	}

	if err != nil {
		Logger.Trace(err)
		return 0, err
	}

	return n, err
}

func (f *file) Write(p []byte) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	f.rwmutex.Lock()
	defer f.rwmutex.Unlock()

	n, err = f.WriteAt(p, f.woffset)
	if err != nil {
		Logger.Trace(err)
		return 0, err
	}

	f.woffset += int64(n)
	return 0, nil
}

func (f *file) WriteAt(p []byte, off int64) (n int, err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return 0, ErrClosed
	}

	f.rwmutex.Lock()
	defer f.rwmutex.Unlock()

	if f.mapped {
		n, err = f.writeMMaps(p, off)
	} else {
		n, err = f.writeFiles(p, off)
	}

	if err != nil {
		Logger.Trace(err)
		return 0, err
	}

	return 0, nil
}

func (f *file) Info() (meta *Metadata) {
	return f.meta
}

func (f *file) Grow(sz int64) (err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return ErrClosed
	}

	f.grmutex.Lock()
	defer f.grmutex.Unlock()

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

func (f *file) MemMap(lock bool) (err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return ErrClosed
	}

	if f.mapped {
		Logger.Error(ErrMapped)
		return nil
	}

	err = f.loadMMaps()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	// complete running tasks before
	// switching read/write mode
	f.rwmutex.Lock()
	defer f.rwmutex.Unlock()

	f.mapped = true
	closeFiles(f.files)

	return nil
}

func (f *file) MUnmap() (err error) {
	if f.closed {
		Logger.Trace(ErrClosed)
		return ErrClosed
	}

	if !f.mapped {
		Logger.Error(ErrNotMapped)
		return nil
	}

	err = f.loadFiles()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	// complete running tasks before
	// switching read/write mode
	f.rwmutex.Lock()
	defer f.rwmutex.Unlock()

	f.mapped = false
	closeMMaps(f.mmaps)

	return nil
}

func (f *file) Close() (err error) {
	if f.closed {
		Logger.Error(ErrClose)
		return nil
	}

	f.rwmutex.Lock()
	defer f.rwmutex.Unlock()
	f.almutex.Lock()
	defer f.almutex.Unlock()

	f.closed = true

	err = f.mdata.Close()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	closeFiles(f.files)
	closeMMaps(f.mmaps)

	return nil
}

// readFiles reads data from files
func (f *file) readFiles(p []byte, off int64) (n int, err error) {
	// TODO code!
	return 0, nil
}

// readMMaps reads data from memory maps
func (f *file) readMMaps(p []byte, off int64) (n int, err error) {
	// TODO code!
	return 0, nil
}

// writeFiles reads data from files
func (f *file) writeFiles(p []byte, off int64) (n int, err error) {
	// TODO code!
	return 0, nil
}

// writeMMaps reads data from memory maps
func (f *file) writeMMaps(p []byte, off int64) (n int, err error) {
	// TODO code!
	return 0, nil
}

// shouldAllocate checks whether there's free space to store sz bytes
func (f *file) shouldAllocate(sz int64) (do bool) {
	meta := f.meta
	total := meta.SegmentSize * meta.SegmentFiles
	return meta.DataSize+sz > total
}

// preallocateIfNeeded pre allocates new segment files if free space
// goes below the threshold (AllocThreshold percentage of SegmentSize).
func (f *file) preallocateIfNeeded() {
	meta := f.meta
	thresh := meta.SegmentSize * AllocThreshold / 100

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
	pathPrefix := path.Join(meta.Directory, meta.FilePrefix)
	filesCount := sz / meta.SegmentSize

	if sz%meta.SegmentSize != 0 {
		filesCount++
	}

	var i int64
	for i = 0; i < filesCount; i++ {
		segID := meta.SegmentFiles + i
		idStr := strconv.Itoa(int(segID))
		fpath := pathPrefix + idStr

		err = createFile(fpath, meta.SegmentSize)
		if err != nil {
			Logger.Trace(err)
			return err
		}

		if f.mapped {
			mfile, err := loadMMap(fpath, meta.SegmentSize)
			if err != nil {
				Logger.Trace(err)
				return err
			}

			f.mmaps = append(f.mmaps, mfile)
		} else {
			file, err := loadFile(fpath, meta.SegmentSize)
			if err != nil {
				Logger.Trace(err)
				return err
			}

			f.files = append(f.files, file)
		}
	}

	meta.SegmentFiles += filesCount

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
	files := make([]*os.File, meta.SegmentFiles)

	if meta.SegmentFiles > 0 {
		var i int64
		for i = 0; i < meta.SegmentFiles; i++ {
			istr := strconv.Itoa(int(i))
			fpath := path.Join(meta.Directory, meta.FilePrefix+istr)

			file, err := loadFile(fpath, meta.SegmentSize)
			if err != nil {
				Logger.Trace(err)
				closeFiles(files)
				return err
			}

			files[i] = file
		}
	}

	prev := f.files
	f.files = files
	closeFiles(prev)

	return nil
}

// loadMMaps loads all segment files as memory maps according to
// segfile metadata. It also ensures all memory maps are valid.
func (f *file) loadMMaps() (err error) {
	meta := f.meta
	mmaps := make([]mmap.File, meta.SegmentFiles)

	if meta.SegmentFiles > 0 {
		var i int64
		for i = 0; i < meta.SegmentFiles; i++ {
			istr := strconv.Itoa(int(i))
			fpath := path.Join(meta.Directory, meta.FilePrefix+istr)

			mfile, err := loadMMap(fpath, meta.SegmentSize)
			if err != nil {
				Logger.Trace(err)
				closeMMaps(mmaps)
				return err
			}

			mmaps[i] = mfile
		}
	}

	prev := f.mmaps
	f.mmaps = mmaps
	closeMMaps(prev)

	return nil
}

// readData reads data from an io.ReaderAt (a file or a memory map)
func readData(r io.ReaderAt, p []byte, off int64) (n int, err error) {
	// TODO code!
	return 0, nil
}

// writeData writes data to a io.WriterAt (a file or a memory map)
func writeData(w io.WriterAt, p []byte, off int64) (n int, err error) {
	// TODO code!
	return 0, nil
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

	return mfile, nil
}

// closeFiles closes a slice of files
func closeFiles(files []*os.File) {
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
func closeMMaps(mmaps []mmap.File) {
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
