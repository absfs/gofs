// Package gofs provides a compatibility layer that adapts absfs.Filer implementations
// to the standard io/fs.FS interface introduced in Go 1.16.
//
// Deprecated: This package is no longer necessary. The absfs package now provides
// native fs.FS support through:
//   - Filer.Sub() returns fs.FS directly (implements fs.SubFS)
//   - Filer.ReadDir() and Filer.ReadFile() provide fs.ReadDirFS/fs.ReadFileFS semantics
//   - absfs.FilerToFS() helper converts any Filer to fs.FS
//   - absfs.File already implements fs.File
//
// Migration guide:
//
//	// Old code:
//	stdFS, _ := gofs.NewFs(myFiler)
//	data, _ := stdFS.ReadFile("hello.txt")
//
//	// New code (option 1 - use Filer.Sub to get fs.FS):
//	stdFS, _ := myFiler.Sub(".")
//	data, _ := fs.ReadFile(stdFS, "hello.txt")
//
//	// New code (option 2 - use FilerToFS helper):
//	stdFS, _ := absfs.FilerToFS(myFiler, ".")
//	data, _ := fs.ReadFile(stdFS, "hello.txt")
//
//	// New code (option 3 - use Filer methods directly):
//	data, _ := myFiler.ReadFile("hello.txt")
//
// This package will be removed in a future release.
package gofs

import (
	"io"
	"io/fs"
	"os"

	"github.com/absfs/absfs"
)

// FileSystem wraps an absfs.Filer to provide compatibility with Go's io/fs interfaces.
// It implements fs.FS, fs.ReadFileFS, fs.ReadDirFS, and fs.StatFS.
//
// Deprecated: Use absfs.FilerToFS() or Filer.Sub(".") instead.
type FileSystem struct {
	Fs absfs.Filer
}

// File wraps an absfs.File to provide compatibility with io/fs.File.
// When the underlying file is a directory, it also implements fs.ReadDirFile.
type File struct {
	F absfs.File
}

// DirEntry wraps an os.FileInfo to provide compatibility with fs.DirEntry.
type DirEntry struct {
	FileInfo os.FileInfo
}

// NewFs creates a new FileSystem that wraps the provided absfs.Filer.
// The returned FileSystem can be used with any Go standard library function
// that accepts fs.FS, fs.ReadFileFS, fs.ReadDirFS, or fs.StatFS.
//
// Deprecated: Use absfs.FilerToFS(filer, ".") or filer.Sub(".") instead.
func NewFs(fs absfs.Filer) (FileSystem, error) {
	return FileSystem{fs}, nil
}

// Open opens the named file for reading and returns it as an fs.File.
// This implements the fs.FS interface.
func (f FileSystem) Open(name string) (fs.File, error) {
	file, err := f.Fs.OpenFile(name, absfs.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return File{file}, nil
}

// ReadDir reads the directory named by name and returns a list of directory entries.
// This implements the fs.ReadDirFS interface.
func (f FileSystem) ReadDir(name string) (dirs []fs.DirEntry, err error) {
	var file absfs.File

	file, err = f.Fs.OpenFile(name, absfs.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			file.Close()
			return
		}
		err = file.Close()
	}()

	var list []os.FileInfo
	list, err = file.Readdir(0)
	if err != nil {
		return nil, err
	}

	dirs = make([]fs.DirEntry, 0, len(list))
	for _, info := range list {
		// Skip . and .. entries - io/fs interface doesn't include them
		if info.Name() == "." || info.Name() == ".." {
			continue
		}
		dirs = append(dirs, DirEntry{info})
	}
	return dirs, nil
}

// ReadFile reads the named file and returns its contents.
// This implements the fs.ReadFileFS interface.
func (f FileSystem) ReadFile(name string) (data []byte, err error) {
	var file absfs.File
	file, err = f.Fs.OpenFile(name, absfs.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			file.Close()
			return
		}
		err = file.Close()
	}()

	return io.ReadAll(file)
}

// Stat returns file information for the named file.
// This implements the fs.StatFS interface.
func (f FileSystem) Stat(name string) (fs.FileInfo, error) {
	return f.Fs.Stat(name)
}

// Sub returns an fs.FS corresponding to the subtree rooted at dir.
// This implements the fs.SubFS interface.
func (f FileSystem) Sub(dir string) (fs.FS, error) {
	return absfs.FilerToFS(f.Fs, dir)
}

// Stat returns file information for this file.
// This implements the fs.File interface.
func (f File) Stat() (fs.FileInfo, error) {
	return f.F.Stat()
}

// Read reads up to len(data) bytes from the file.
// This implements the io.Reader interface.
func (f File) Read(data []byte) (int, error) {
	return f.F.Read(data)
}

// ReadDir reads the contents of the directory associated with the file f
// and returns a slice of DirEntry values in directory order.
// If n > 0, ReadDir returns at most n DirEntry structures.
// If n <= 0, ReadDir returns all the DirEntry values from the directory in a single slice.
// This implements the fs.ReadDirFile interface.
func (f File) ReadDir(n int) (dirs []fs.DirEntry, err error) {
	// If n <= 0, read all entries at once
	if n <= 0 {
		var list []os.FileInfo
		list, err = f.F.Readdir(0)
		if err != nil {
			return nil, err
		}

		dirs = make([]fs.DirEntry, 0, len(list))
		for _, info := range list {
			// Skip . and .. entries - io/fs interface doesn't include them
			if info.Name() == "." || info.Name() == ".." {
				continue
			}
			dirs = append(dirs, DirEntry{info})
		}
		return dirs, nil
	}

	// n > 0: read until we have n valid entries (excluding . and ..)
	dirs = make([]fs.DirEntry, 0, n)
	for len(dirs) < n {
		// Read one entry at a time to ensure we get exactly n valid entries
		list, err := f.F.Readdir(1)
		if err != nil {
			if err == io.EOF && len(dirs) > 0 {
				// Return what we have so far
				return dirs, nil
			}
			return dirs, err
		}

		if len(list) == 0 {
			break
		}

		// Skip . and .. entries
		if list[0].Name() != "." && list[0].Name() != ".." {
			dirs = append(dirs, DirEntry{list[0]})
		}
	}

	return dirs, nil
}

// Close closes the file, rendering it unusable for I/O.
// This implements the fs.File interface.
func (f File) Close() error {
	return f.F.Close()
}

// Name returns the base name of the file.
func (d DirEntry) Name() string {
	return d.FileInfo.Name()
}

// IsDir reports whether the entry describes a directory.
func (d DirEntry) IsDir() bool {
	return d.FileInfo.IsDir()
}

// Type returns the type bits for the entry.
func (d DirEntry) Type() fs.FileMode {
	return d.FileInfo.Mode()
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
func (d DirEntry) Info() (fs.FileInfo, error) {
	return d.FileInfo, nil
}
