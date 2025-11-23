package gofs

import (
	"io"
	"testing"

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

// BenchmarkNewFs benchmarks creating a new FileSystem wrapper.
func BenchmarkNewFs(b *testing.B) {
	mfs, err := memfs.NewFS()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewFs(mfs)
	}
}

// BenchmarkOpen benchmarks opening files.
func BenchmarkOpen(b *testing.B) {
	mfs, _ := memfs.NewFS()
	file, _ := mfs.Create("benchmark.txt")
	file.Write([]byte("benchmark data"))
	file.Close()

	gfs, _ := NewFs(mfs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := gfs.Open("benchmark.txt")
		f.Close()
	}
}

// BenchmarkReadFile benchmarks reading entire files.
func BenchmarkReadFile(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_1KB", 1024},
		{"Medium_64KB", 64 * 1024},
		{"Large_1MB", 1024 * 1024},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			mfs, _ := memfs.NewFS()
			data := make([]byte, sz.size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			writeFile(mfs, "test.dat", data)

			gfs, _ := NewFs(mfs)

			b.ResetTimer()
			b.SetBytes(int64(sz.size))
			for i := 0; i < b.N; i++ {
				_, _ = gfs.ReadFile("test.dat")
			}
		})
	}
}

// BenchmarkReadDir benchmarks reading directory contents.
func BenchmarkReadDir(b *testing.B) {
	dirSizes := []struct {
		name  string
		files int
	}{
		{"Small_10", 10},
		{"Medium_100", 100},
		{"Large_1000", 1000},
	}

	for _, ds := range dirSizes {
		b.Run(ds.name, func(b *testing.B) {
			mfs, _ := memfs.NewFS()
			mfs.Mkdir("bench", 0755)

			for i := 0; i < ds.files; i++ {
				name := "bench/file" + string(rune('0'+i%10)) + ".txt"
				writeFile(mfs, name, []byte("data"))
			}

			gfs, _ := NewFs(mfs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = gfs.ReadDir("bench")
			}
		})
	}
}

// BenchmarkStat benchmarks getting file information.
func BenchmarkStat(b *testing.B) {
	mfs, _ := memfs.NewFS()
	writeFile(mfs, "stat.txt", []byte("data"))

	gfs, _ := NewFs(mfs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gfs.Stat("stat.txt")
	}
}

// BenchmarkFileRead benchmarks reading from an open file.
func BenchmarkFileRead(b *testing.B) {
	mfs, _ := memfs.NewFS()
	data := make([]byte, 64*1024) // 64KB
	for i := range data {
		data[i] = byte(i % 256)
	}
	writeFile(mfs, "read.dat", data)

	gfs, _ := NewFs(mfs)

	b.Run("Sequential", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			f, _ := gfs.Open("read.dat")
			buf := make([]byte, 4096)
			for {
				_, err := f.Read(buf)
				if err != nil {
					break
				}
			}
			f.Close()
		}
	})

	b.Run("FullRead", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		for i := 0; i < b.N; i++ {
			f, _ := gfs.Open("read.dat")
			io.ReadAll(f)
			f.Close()
		}
	})
}

// BenchmarkFileReadDir benchmarks reading directory entries from an open file.
func BenchmarkFileReadDir(b *testing.B) {
	mfs, _ := memfs.NewFS()
	mfs.Mkdir("readdir", 0755)

	for i := 0; i < 100; i++ {
		name := "readdir/file" + string(rune('0'+i%10)) + ".txt"
		writeFile(mfs, name, []byte("data"))
	}

	gfs, _ := NewFs(mfs)

	b.Run("ReadAll", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			f, _ := gfs.Open("readdir")
			file := f.(File)
			_, _ = file.ReadDir(0)
			f.Close()
		}
	})

	b.Run("ReadLimited", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			f, _ := gfs.Open("readdir")
			file := f.(File)
			_, _ = file.ReadDir(10)
			f.Close()
		}
	})
}

// BenchmarkDirEntryOps benchmarks DirEntry operations.
func BenchmarkDirEntryOps(b *testing.B) {
	mfs, _ := memfs.NewFS()
	mfs.Mkdir("entries", 0755)
	writeFile(mfs, "entries/test.txt", []byte("data"))

	gfs, _ := NewFs(mfs)
	entries, _ := gfs.ReadDir("entries")

	if len(entries) == 0 {
		b.Fatal("No entries found")
	}

	// Find a non-dot entry
	var entry DirEntry
	for _, e := range entries {
		if e.Name() != "." && e.Name() != ".." {
			entry = e.(DirEntry)
			break
		}
	}

	b.Run("Name", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = entry.Name()
		}
	})

	b.Run("IsDir", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = entry.IsDir()
		}
	})

	b.Run("Type", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = entry.Type()
		}
	})

	b.Run("Info", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = entry.Info()
		}
	})
}

// BenchmarkConcurrentReads benchmarks concurrent file reads.
func BenchmarkConcurrentReads(b *testing.B) {
	mfs, _ := memfs.NewFS()
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	writeFile(mfs, "concurrent.dat", data)

	gfs, _ := NewFs(mfs)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = gfs.ReadFile("concurrent.dat")
		}
	})
}
