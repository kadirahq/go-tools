package fsutils

import (
	"errors"
	"io"
	"os"
	"path"

	goerr "github.com/go-errors/errors"
)

var (
	// ChunkSize is the number of bytes written at a time
	ChunkSize = int64(1024 * 1024)

	// ChunkData is a slice of zero bytes of size ChunkSize
	ChunkData = make([]byte, ChunkSize)

	// ErrEmptyPath is returned when given file/dir path is an empty string
	ErrEmptyPath = errors.New("cannot use empty string as path")

	// ErrWriteSz is returned when bytes written is not equal to data size
	ErrWriteSz = errors.New("bytes written != data size")

	// ErrReadSz is returned when bytes read is not equal to data size
	ErrReadSz = errors.New("bytes read != data size")

	// ErrFileDir is returned when a file was found instead of a directory
	ErrFileDir = errors.New("expecting a directory, got a file")

	// ErrDirFile is returned when a directory was found instead of a file
	ErrDirFile = errors.New("expecting a file, got a directory")
)

// Syncer does some sort of a synchronization action. It may be writing
// data to disk, synchronize data over network or something else.
type Syncer interface {
	Sync() (err error)
}

// EnsureDir makes sure that a directory exists at path
// It will attempt to create a directory if not exists
func EnsureDir(dpath string) (err error) {
	if dpath == "" {
		return goerr.Wrap(ErrEmptyPath, 0)
	}

	if err := os.MkdirAll(dpath, 0755); err != nil {
		return goerr.Wrap(err, 0)
	}

	return nil
}

// EnsureFile makes sure that a file exists at path with at least given size.
// If the file size is smaller, empty bytes will be appended at file end.
// It returns the file handler, final size and an error (if any).
// NOTE: Always get the file handler and close it after use.
func EnsureFile(fpath string, sz int64) (size int64, err error) {
	if fpath == "" {
		return 0, goerr.Wrap(ErrEmptyPath, 0)
	}

	dpath := path.Dir(fpath)
	if err := EnsureDir(dpath); err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	file, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	defer file.Close()

	finfo, err := file.Stat()
	if err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	if finfo.IsDir() {
		return 0, goerr.Wrap(ErrDirFile, 0)
	}

	filesize := finfo.Size()
	if filesize >= sz {
		return filesize, nil
	}

	if err = ZeroFill(file, sz-filesize, filesize); err != nil {
		return 0, goerr.Wrap(err, 0)
	}

	return sz, nil
}

// ZeroFill writes given number of zeroes to the given file.
func ZeroFill(file io.WriterAt, sz, off int64) (err error) {
	chunks := sz / ChunkSize

	var i int64
	for i = 0; i < chunks; i++ {
		offset := off + ChunkSize*i
		n, err := file.WriteAt(ChunkData, offset)
		if err != nil {
			return err
		} else if int64(n) != ChunkSize {
			return goerr.Wrap(ErrWriteSz, 0)
		}
	}

	// write all remaining bytes
	toWrite := sz % ChunkSize
	zeroes := ChunkData[:toWrite]
	offset := ChunkSize * chunks
	n, err := file.WriteAt(zeroes, offset)
	if err != nil {
		return err
	} else if int64(n) != toWrite {
		return goerr.Wrap(ErrWriteSz, 0)
	}

	return nil
}
