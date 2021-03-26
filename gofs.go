package gofs

import (
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/absfs/absfs"
)

type FileSystem struct {
	Fs absfs.Filer
}

type File struct {
	F absfs.File
}

type DirEntry struct {
	FileInfo os.FileInfo
}

func NewFs(fs absfs.Filer) (FileSystem, error) {
	return FileSystem{fs}, nil
}

func (f FileSystem) Open(name string) (fs.File, error) {
	file, err := f.Fs.OpenFile(name, absfs.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return File{file}, nil
}

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

	return ioutil.ReadAll(file)
}

func (f FileSystem) Stat(name string) (fs.FileInfo, error) {
	return f.Fs.Stat(name)
}

func (f File) Stat() (fs.FileInfo, error) {
	return f.F.Stat()
}

func (f File) Read(data []byte) (int, error) {
	return f.F.Read(data)
}

func (f File) ReadDir(n int) (dirs []fs.DirEntry, err error) {
	var list []os.FileInfo
	list, err = f.F.Readdir(0)
	if err != nil {
		return nil, err
	}

	dirs = make([]fs.DirEntry, 0, len(list))
	for _, info := range list {
		dirs = append(dirs, DirEntry{info})
	}
	return dirs, nil
}

func (f File) Close() error {
	return f.F.Close()
}

func (d DirEntry) Name() string {
	return d.FileInfo.Name()
}

func (d DirEntry) IsDir() bool {
	return d.FileInfo.IsDir()
}

func (d DirEntry) Type() fs.FileMode {
	return d.FileInfo.Mode()
}

func (d DirEntry) Info() (fs.FileInfo, error) {
	return d.FileInfo, nil
}
