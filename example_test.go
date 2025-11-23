package gofs_test

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/absfs/gofs"
	"github.com/absfs/memfs"
)

// Helper function to write a file to memfs
func writeFile(fs *memfs.FileSystem, name string, data []byte) error {
	f, err := fs.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// Example demonstrates basic usage of gofs to wrap an absfs filesystem
// and use it with Go's standard library functions.
func Example() {
	// Create an in-memory filesystem
	mfs, err := memfs.NewFS()
	if err != nil {
		log.Fatal(err)
	}

	// Create a test file
	file, _ := mfs.Create("hello.txt")
	file.Write([]byte("Hello, World!"))
	file.Close()

	// Wrap it with gofs for io/fs compatibility
	stdFS, err := gofs.NewFs(mfs)
	if err != nil {
		log.Fatal(err)
	}

	// Read the file using the standard interface
	data, err := stdFS.ReadFile("hello.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Output: Hello, World!
}

// ExampleNewFs demonstrates creating a gofs filesystem wrapper.
func ExampleNewFs() {
	// Create any absfs filesystem implementation
	mfs, _ := memfs.NewFS()

	// Wrap it with gofs
	fsys, err := gofs.NewFs(mfs)
	if err != nil {
		log.Fatal(err)
	}

	// Now fsys can be used with any Go function accepting fs.FS
	_ = fsys
	fmt.Println("FileSystem created successfully")
	// Output: FileSystem created successfully
}

// ExampleFileSystem_Open demonstrates opening a file.
func ExampleFileSystem_Open() {
	mfs, _ := memfs.NewFS()
	file, _ := mfs.Create("example.txt")
	file.Write([]byte("content"))
	file.Close()

	fsys, _ := gofs.NewFs(mfs)

	// Open the file
	f, err := fsys.Open("example.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Read from the file
	buf := make([]byte, 7)
	n, _ := f.Read(buf)
	fmt.Printf("Read %d bytes: %s\n", n, string(buf))
	// Output: Read 7 bytes: content
}

// ExampleFileSystem_ReadFile demonstrates reading an entire file.
func ExampleFileSystem_ReadFile() {
	mfs, _ := memfs.NewFS()
	writeFile(mfs, "data.txt", []byte("file contents"))

	fsys, _ := gofs.NewFs(mfs)

	// Read the entire file
	data, err := fsys.ReadFile("data.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Output: file contents
}

// ExampleFileSystem_ReadDir demonstrates reading directory contents.
func ExampleFileSystem_ReadDir() {
	mfs, _ := memfs.NewFS()
	mfs.Mkdir("mydir", 0755)
	writeFile(mfs, "mydir/file1.txt", []byte("one"))
	writeFile(mfs, "mydir/file2.txt", []byte("two"))

	fsys, _ := gofs.NewFs(mfs)

	// Read directory contents
	entries, err := fsys.ReadDir("mydir")
	if err != nil {
		log.Fatal(err)
	}

	// Print file names (excluding . and ..)
	for _, entry := range entries {
		if entry.Name() != "." && entry.Name() != ".." {
			fmt.Println(entry.Name())
		}
	}
	// Output:
	// file1.txt
	// file2.txt
}

// ExampleFileSystem_Stat demonstrates getting file information.
func ExampleFileSystem_Stat() {
	mfs, _ := memfs.NewFS()
	writeFile(mfs, "info.txt", []byte("test data"))

	fsys, _ := gofs.NewFs(mfs)

	// Get file info
	info, err := fsys.Stat("info.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s\n", info.Name())
	fmt.Printf("Size: %d bytes\n", info.Size())
	fmt.Printf("IsDir: %v\n", info.IsDir())
	// Output:
	// Name: info.txt
	// Size: 9 bytes
	// IsDir: false
}

// Example_walkDir demonstrates using gofs with fs.WalkDir.
func Example_walkDir() {
	mfs, _ := memfs.NewFS()
	mfs.Mkdir("project", 0755)
	writeFile(mfs, "project/main.go", []byte("package main"))
	writeFile(mfs, "project/README.md", []byte("# Project"))

	fsys, _ := gofs.NewFs(mfs)

	// Walk the filesystem
	fs.WalkDir(fsys, "project/main.go", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() != "." && d.Name() != ".." {
			fmt.Printf("%s (dir: %v)\n", d.Name(), d.IsDir())
		}
		return nil
	})
	// Output: main.go (dir: false)
}

// Example_glob demonstrates using gofs with fs.Glob.
func Example_glob() {
	mfs, _ := memfs.NewFS()
	writeFile(mfs, "test1.txt", []byte("one"))
	writeFile(mfs, "test2.txt", []byte("two"))
	writeFile(mfs, "readme.md", []byte("doc"))

	fsys, _ := gofs.NewFs(mfs)

	// Find all .txt files
	matches, err := fs.Glob(fsys, "*.txt")
	if err != nil {
		log.Fatal(err)
	}

	for _, match := range matches {
		fmt.Println(match)
	}
	// Output:
	// test1.txt
	// test2.txt
}

// Example_readDir demonstrates using gofs with fs.ReadDir.
func Example_readDir() {
	mfs, _ := memfs.NewFS()
	mfs.Mkdir("docs", 0755)
	writeFile(mfs, "docs/api.md", []byte("API"))
	writeFile(mfs, "docs/guide.md", []byte("Guide"))

	fsys, _ := gofs.NewFs(mfs)

	// Read directory using standard library function
	entries, err := fs.ReadDir(fsys, "docs")
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Name() != "." && entry.Name() != ".." {
			fmt.Printf("%s: %d bytes\n", entry.Name(), mustSize(entry))
		}
	}
	// Output:
	// api.md: 3 bytes
	// guide.md: 5 bytes
}

func mustSize(entry fs.DirEntry) int64 {
	info, err := entry.Info()
	if err != nil {
		return 0
	}
	return info.Size()
}

// Example_fileRead demonstrates reading a file incrementally.
func Example_fileRead() {
	mfs, _ := memfs.NewFS()
	writeFile(mfs, "stream.txt", []byte("ABCDEFGH"))

	fsys, _ := gofs.NewFs(mfs)

	f, _ := fsys.Open("stream.txt")
	defer f.Close()

	// Read in chunks
	buf := make([]byte, 3)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			fmt.Println(string(buf[:n]))
		}
		if err != nil {
			break
		}
	}
	// Output:
	// ABC
	// DEF
	// GH
}
