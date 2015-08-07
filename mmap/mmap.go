package mmap

import (
	"errors"
	"io"
	"os"
	"path"
	"sync"
	"syscall"

	"github.com/kadirahq/go-tools/logger"
)

const (
	// DirectoryPerm is the permission set for new directories
	DirectoryPerm = 0755

	// FileMode used when opening files for memory mapping
	FileMode = os.O_CREATE | os.O_RDWR

	// FilePerm is the permissions used when creating new files
	FilePerm = 0644

	// FileProt is the memory map prot parameter
	FileProt = syscall.PROT_READ | syscall.PROT_WRITE

	// FileFlag is the memory map flag parameter
	FileFlag = syscall.MAP_SHARED

	// ChunkSize is the number of bytes to write at a time.
	// When creating new files, create it in small chunks.
	ChunkSize = 1024 * 1024 * 10
)

var (
	// ErrWrite is returned when bytes written not equal to data size
	ErrWrite = errors.New("bytes written != data size")

	// ErrOptions is returned when options have missing or invalid fields.
	ErrOptions = errors.New("invalid or missing options")

	// ErrFileDir is returned when file at path is a directory
	ErrFileDir = errors.New("directory at target path")

	// ChunkBytes is a ChunkSize size slice of zeroes
	ChunkBytes = make([]byte, ChunkSize)

	// Logger logs stuff
	Logger = logger.New("MMAP")
)

// Options has parameters required for creating an `Index`
type Options struct {
	// memory map file path
	// this field if required
	Path string

	// minimum size of the mmap file
	// if not provided, file size will be used
	Size int64

	// TODO support mapping only a part of a file
	// offset to start the memory map from
	// Offset int64
}

// File is similar to os.File but all reads/writes are done through
// memory maps. This can often lead to much faster reads/writes.
type File interface {
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt

	// Size returns the size of the memory map
	Size() (sz int64)

	// Reset sets io.Reader, io.Writer offsets to zero
	// TODO check whether we need separate reset functions
	Reset()

	// Grow method grows the file by `size` number of bytes.
	// Once it's done, the file will be re-mapped with added bytes.
	Grow(size int64) (err error)

	// Lock loads memory mapped data to the RAM and keeps them in RAM.
	// If not done, the data will be kept on disk until required.
	// Locking a memory map can decrease initial page faults.
	Lock() (err error)

	// Unlock releases memory by not reserving parts of RAM of the file.
	// The OS may use memory mapped data from the disk when done.
	Unlock() (err error)

	// Close method unmaps data and closes the file.
	// If the mmap is locked, it'll be unlocked first.
	Close() (err error)
}

type mfile struct {
	// options used when creating the memory map
	options *Options

	// byte slice which contains memory mapped data
	data []byte

	// current size of the memory mapped data
	size int64

	// map file handler used when growing the map
	file *os.File

	// indicates whether the memory map is locked or not
	lock bool

	// read/write mutex
	rwmutx *sync.RWMutex

	// growth mutex to control Grow calls
	grmutx *sync.Mutex

	// io.Reader read offset
	roffset int64

	// io.Reader write offset
	woffset int64
}

// New creates a File struct with given options.
// Default values will be used for missing options.
func New(options *Options) (mf File, err error) {
	// validate options
	if options == nil ||
		options.Path == "" {
		Logger.Trace(ErrOptions)
		return nil, ErrOptions
	}

	dpath := path.Dir(options.Path)
	err = os.MkdirAll(dpath, DirectoryPerm)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	file, err := os.OpenFile(options.Path, FileMode, FilePerm)
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
		Logger.Trace(ErrFileDir)
		return nil, ErrFileDir
	}

	size := finfo.Size()
	if options.Size == 0 {
		options.Size = size
	}

	if toGrow := options.Size - size; toGrow > 0 {
		err = growFile(file, toGrow)
		if err != nil {
			Logger.Trace(err)
			return nil, err
		}

		size = options.Size
	}

	data, err := mmapFile(file, 0, size)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	mf = &mfile{
		options: options,
		data:    data,
		size:    size,
		file:    file,
		rwmutx:  &sync.RWMutex{},
		grmutx:  &sync.Mutex{},
	}

	return mf, nil
}

func (m *mfile) Read(p []byte) (n int, err error) {
	m.rwmutx.RLock()
	defer m.rwmutx.RUnlock()

	n, err = m.read(p, m.roffset)
	if err == nil {
		m.roffset += int64(n)
	} else {
		Logger.Trace(err)
	}

	return n, err
}

func (m *mfile) ReadAt(p []byte, off int64) (n int, err error) {
	m.rwmutx.RLock()
	defer m.rwmutx.RUnlock()
	return m.read(p, off)
}

func (m *mfile) Write(p []byte) (n int, err error) {
	m.rwmutx.Lock()
	defer m.rwmutx.Unlock()

	n, err = m.write(p, m.woffset)
	if err == nil {
		m.woffset += int64(n)
	} else {
		Logger.Trace(err)
	}

	return n, err
}

func (m *mfile) WriteAt(p []byte, off int64) (n int, err error) {
	m.rwmutx.Lock()
	defer m.rwmutx.Unlock()
	return m.write(p, off)
}

func (m *mfile) Size() (sz int64) {
	m.rwmutx.RLock()
	defer m.rwmutx.RUnlock()
	return m.size
}

func (m *mfile) Reset() {
	m.rwmutx.Lock()
	defer m.rwmutx.Unlock()

	m.roffset = 0
	m.woffset = 0
}

func (m *mfile) Grow(size int64) (err error) {
	m.rwmutx.Lock()
	defer m.rwmutx.Unlock()
	return m.grow(size)
}

func (m *mfile) Lock() (err error) {
	if m.lock {
		return nil
	}

	err = lockData(m.data)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	m.lock = true
	return nil
}

func (m *mfile) Unlock() (err error) {
	if !m.lock {
		return nil
	}

	err = unlockData(m.data)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	m.lock = false
	return nil
}

func (m *mfile) Close() (err error) {
	err = m.Unlock()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	m.rwmutx.Lock()
	defer m.rwmutx.Unlock()

	err = unmapData(m.data)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	err = m.file.Close()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	return nil
}

func (m *mfile) read(p []byte, off int64) (n int, err error) {
	var src []byte
	var end = off + int64(len(p))

	if end > m.size {
		err = io.EOF
		src = m.data[off:m.size]
		n = int(m.size - off)
	} else {
		src = m.data[off:end]
		n = int(end - off)
	}

	copy(p, src)
	return n, err
}

func (m *mfile) write(p []byte, off int64) (n int, err error) {
	var dst []byte
	var end = off + int64(len(p))

	if end > m.size {
		toGrow := end - m.size
		err = m.grow(toGrow)
		if err != nil {
			return 0, err
		}
	}

	dst = m.data[off:end]
	n = int(end - off)
	copy(dst, p)
	return n, nil
}

func (m *mfile) grow(size int64) (err error) {
	lock := m.lock

	if lock {
		err := m.Unlock()
		if err != nil {
			Logger.Trace(err)
			return err
		}

		m.lock = false
	}

	err = unmapData(m.data)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	err = growFile(m.file, size)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	m.size += size
	m.data, err = mmapFile(m.file, 0, m.size)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	if lock {
		err := m.Lock()
		if err != nil {
			Logger.Trace(err)
			return err
		}

		m.lock = true
	}

	return nil
}

// growFile grows a file with `size` number of bytes.
// `fsize` is the current file size in bytes.
// empty bytes are appended to the end of the file.
func growFile(file *os.File, size int64) (err error) {
	finfo, err := file.Stat()
	if err != nil {
		Logger.Trace(err)
		return err
	}

	fsize := finfo.Size()

	// number of complete chunks to write
	chunksCount := size / ChunkSize

	var i int64
	for i = 0; i < chunksCount; i++ {
		offset := fsize + ChunkSize*i
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
	toWrite := size % ChunkSize
	zeroes := ChunkBytes[:toWrite]
	offset := fsize + ChunkSize*chunksCount
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

// mmapFile creates a new memory map with the given file.
// if the file size is zero, a memory cannot be created therefore
// an empty byte array is returned instead (no errors returned).
func mmapFile(file *os.File, from, to int64) (data []byte, err error) {
	fd := int(file.Fd())
	ln := int(to - from)

	if ln == 0 {
		data = make([]byte, 0, 0)
		return data, nil
	}

	data, err = syscall.Mmap(fd, from, ln, FileProt, FileFlag)
	if err != nil {
		Logger.Trace(err)
		return nil, err
	}

	return data, nil
}

// unmapData unmaps mapped data and releases memory
// If the data size is zero, a map cannot exist
// therefore assume no errors and return nil
func unmapData(data []byte) (err error) {
	if len(data) == 0 {
		return nil
	}

	err = syscall.Munmap(data)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	return nil
}

// lockData locks data to physical memory
func lockData(data []byte) (err error) {
	err = syscall.Mlock(data)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	return nil
}

// unlockData releases locked memory
func unlockData(data []byte) (err error) {
	err = syscall.Munlock(data)
	if err != nil {
		Logger.Trace(err)
		return err
	}

	return nil
}
