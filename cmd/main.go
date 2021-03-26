package main

import (
	"fmt"
	"io/fs"

	"github.com/absfs/gofs"
	"github.com/absfs/memfs"
)

func main() {
	mfs, _ := memfs.NewFS()

	memFile, _ := mfs.Create("foo.txt")
	fmt.Fprintf(memFile, "bar\n")
	memFile.Close()

	var fs fs.ReadFileFS
	fs, _ = gofs.NewFs(mfs)

	data, _ := fs.ReadFile("foo.txt")
	fmt.Print(string(data))
	// output: "bar\n"
}
