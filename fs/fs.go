package fs

// Syncer interface provides a Sync method which will
// permenantly store data to a device (eg. hard-disk)
type Syncer interface {
	Sync() (err error)
}

// SlicerAt interface provides an SliceAt method which will
// fetch a slice of bytes from an in-memory data storage.
type SlicerAt interface {
	SliceAt(sz, off int64) (p []byte, err error)
}

// Slicer interface provides a Slice method which will
// fetch a slice of bytes from an in-memory data storage.
type Slicer interface {
	Slice(sz int64) (p []byte, err error)
}

// Ensurer interface provides an Ensure method which will
// make sure that the given offset exists and has a value.
type Ensurer interface {
	Ensure(off int64) (err error)
}
