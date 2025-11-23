// Package gofs provides a compatibility layer that adapts absfs.Filer implementations
// to the standard io/fs.FS interface introduced in Go 1.16.
//
// This package allows any filesystem implementation from the absfs ecosystem
// (such as memfs, osfs, or custom implementations) to be used with Go's standard
// library functions that accept fs.FS, fs.ReadFileFS, fs.ReadDirFS, or fs.StatFS.
//
// The main type is FileSystem, which wraps an absfs.Filer and implements multiple
// standard filesystem interfaces. Files opened through this filesystem implement
// fs.File and, when appropriate, fs.ReadDirFile for directory operations.
//
// Example usage:
//
//	import (
//		"github.com/absfs/gofs"
//		"github.com/absfs/memfs"
//	)
//
//	// Create an in-memory filesystem
//	mfs, _ := memfs.NewFS()
//	f, _ := mfs.Create("hello.txt")
//	f.Write([]byte("Hello, World!"))
//	f.Close()
//
//	// Wrap it with gofs for io/fs compatibility
//	stdFS, _ := gofs.NewFs(mfs)
//
//	// Use with standard library functions
//	data, _ := stdFS.ReadFile("hello.txt")
//	fmt.Println(string(data)) // Output: Hello, World!
package gofs

import (
	"io"
	"io/fs"
	"os"

	"github.com/absfs/absfs"
)

// FileSystem wraps an absfs.Filer to provide compatibility with Go's io/fs interfaces.
// It implements fs.FS, fs.ReadFileFS, fs.ReadDirFS, and fs.StatFS.
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
	var list []os.FileInfo
	list, err = f.F.Readdir(n)
	if err != nil {
		return nil, err
	}

	dirs = make([]fs.DirEntry, 0, len(list))
	for _, info := range list {
		dirs = append(dirs, DirEntry{info})
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
