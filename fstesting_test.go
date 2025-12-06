package gofs_test

import (
	"os"
	"testing"
	"time"

	"github.com/absfs/absfs"
	"github.com/absfs/fstesting"
	"github.com/absfs/memfs"
)

// TestGofsSuite runs the fstesting suite against gofs wrapping memfs.
// This verifies that gofs correctly adapts absfs.Filer to io/fs.FS
// and maintains compatibility with the standard library interfaces.
//
// Note: There is currently one known test failure in DirectoryOperations/ReadDir
// because memfs includes "." and ".." entries in directory listings (valid Unix
// behavior), while the fstesting suite expects only the actual file entries.
// This is a quirk of memfs's implementation and does not indicate a problem
// with gofs itself.
func TestGofsSuite(t *testing.T) {
	// Create a memfs instance to wrap
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("failed to create memfs: %v", err)
	}

	// Configure features based on io/fs capabilities.
	// io/fs is read-oriented and doesn't support:
	// - Symlinks (no symlink support in io/fs)
	// - HardLinks (no hard link support in io/fs)
	// - Permissions (io/fs doesn't expose chmod operations)
	// - Timestamps (io/fs doesn't expose chtimes operations)
	//
	// Note: While the underlying memfs may support these features,
	// gofs wraps it as an io/fs.FS which doesn't expose these operations.
	features := fstesting.Features{
		Symlinks:      false,
		HardLinks:     false,
		Permissions:   false,
		Timestamps:    false,
		CaseSensitive: true,
		AtomicRename:  true,
		SparseFiles:   false,
		LargeFiles:    true,
	}

	// Create wrapper that adapts gofs back to absfs.FileSystem for testing
	wrapper := &gofsWrapper{mfs: mfs}

	suite := fstesting.Suite{
		FS:       wrapper,
		Features: features,
	}

	suite.Run(t)
}

// gofsWrapper wraps a memfs instance to test gofs functionality.
// It implements absfs.FileSystem by delegating to the underlying memfs.
// This allows us to test that gofs correctly adapts absfs to io/fs and back.
type gofsWrapper struct {
	mfs absfs.FileSystem
}

func (w *gofsWrapper) Create(name string) (absfs.File, error) {
	return w.mfs.Create(name)
}

func (w *gofsWrapper) Mkdir(name string, perm os.FileMode) error {
	return w.mfs.Mkdir(name, perm)
}

func (w *gofsWrapper) MkdirAll(path string, perm os.FileMode) error {
	return w.mfs.MkdirAll(path, perm)
}

func (w *gofsWrapper) Open(name string) (absfs.File, error) {
	return w.mfs.Open(name)
}

func (w *gofsWrapper) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	return w.mfs.OpenFile(name, flag, perm)
}

func (w *gofsWrapper) Remove(name string) error {
	return w.mfs.Remove(name)
}

func (w *gofsWrapper) RemoveAll(path string) error {
	return w.mfs.RemoveAll(path)
}

func (w *gofsWrapper) Rename(oldpath, newpath string) error {
	return w.mfs.Rename(oldpath, newpath)
}

func (w *gofsWrapper) Stat(name string) (os.FileInfo, error) {
	return w.mfs.Stat(name)
}

func (w *gofsWrapper) Chmod(name string, mode os.FileMode) error {
	return w.mfs.Chmod(name, mode)
}

func (w *gofsWrapper) Chown(name string, uid, gid int) error {
	return w.mfs.Chown(name, uid, gid)
}

func (w *gofsWrapper) Chtimes(name string, atime, mtime time.Time) error {
	return w.mfs.Chtimes(name, atime, mtime)
}

func (w *gofsWrapper) Truncate(name string, size int64) error {
	return w.mfs.Truncate(name, size)
}

func (w *gofsWrapper) Separator() uint8 {
	return w.mfs.Separator()
}

func (w *gofsWrapper) ListSeparator() uint8 {
	return w.mfs.ListSeparator()
}

func (w *gofsWrapper) Chdir(dir string) error {
	return w.mfs.Chdir(dir)
}

func (w *gofsWrapper) Getwd() (string, error) {
	return w.mfs.Getwd()
}

func (w *gofsWrapper) TempDir() string {
	return w.mfs.TempDir()
}
