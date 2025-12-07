# gofs - Abstract File System interface

[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/gofs.svg)](https://pkg.go.dev/github.com/absfs/gofs)
[![Go Report Card](https://goreportcard.com/badge/github.com/absfs/gofs)](https://goreportcard.com/report/github.com/absfs/gofs)
[![CI](https://github.com/absfs/gofs/actions/workflows/ci.yml/badge.svg)](https://github.com/absfs/gofs/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

The `gofs` package adds Go `fs` filesystem interface support to any [`absfs`][1]
filesystem implementation.

## Import

```go
import "github.com/absfs/gofs"
```

## Example Usage

```go
package main

import (
	"fmt"

	"github.com/absfs/gofs"
	"github.com/absfs/memfs"
)

func main() {
	mfs, _ := memfs.NewFS()

	memFile, _ := mfs.Create("foo.txt")
	fmt.Fprintf(memFile, "bar\n")
	memFile.Close()

	var fs gofs.FileSystem
	fs, _ = gofs.NewFs(mfs)

	data, _ := fs.ReadFile("foo.txt")
	fmt.Print(string(data))
	// output: "bar\n"
}

```

## absfs
Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract filesystem interface.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/gofs/blob/master/LICENSE)


[1]:(https://github.com/absfs/absfs)
