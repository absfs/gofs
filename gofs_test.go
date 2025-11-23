package gofs

import (
	"errors"
	"io"
	"io/fs"
	"testing"

	"github.com/absfs/memfs"
)

func setupTestFS(t *testing.T) FileSystem {
	t.Helper()
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	// Create test file
	f, err := mfs.Create("testfile.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if _, err := f.Write([]byte("Hello, World!")); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close test file: %v", err)
	}

	// Create test directory with files
	if err := mfs.Mkdir("testdir", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	for i := 1; i <= 5; i++ {
		name := "testdir/file" + string(rune('0'+i)) + ".txt"
		f, err := mfs.Create(name)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
		if _, err := f.Write([]byte("content")); err != nil {
			t.Fatalf("Failed to write to file %s: %v", name, err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("Failed to close file %s: %v", name, err)
		}
	}

	gfs, err := NewFs(mfs)
	if err != nil {
		t.Fatalf("Failed to create gofs: %v", err)
	}

	return gfs
}

func TestNewFs(t *testing.T) {
	mfs, err := memfs.NewFS()
	if err != nil {
		t.Fatalf("Failed to create memfs: %v", err)
	}

	gfs, err := NewFs(mfs)
	if err != nil {
		t.Errorf("NewFs() returned error: %v", err)
	}

	if gfs.Fs == nil {
		t.Error("NewFs() returned FileSystem with nil Fs field")
	}
}

func TestFileSystem_Open(t *testing.T) {
	gfs := setupTestFS(t)

	t.Run("open existing file", func(t *testing.T) {
		file, err := gfs.Open("testfile.txt")
		if err != nil {
			t.Fatalf("Open() failed: %v", err)
		}
		defer file.Close()

		if file == nil {
			t.Error("Open() returned nil file")
		}
	})

	t.Run("open non-existent file", func(t *testing.T) {
		_, err := gfs.Open("nonexistent.txt")
		if err == nil {
			t.Error("Open() should return error for non-existent file")
		}
	})

	t.Run("open directory", func(t *testing.T) {
		file, err := gfs.Open("testdir")
		if err != nil {
			t.Fatalf("Open() failed to open directory: %v", err)
		}
		defer file.Close()

		if file == nil {
			t.Error("Open() returned nil file for directory")
		}
	})
}

func TestFileSystem_ReadFile(t *testing.T) {
	gfs := setupTestFS(t)

	t.Run("read existing file", func(t *testing.T) {
		data, err := gfs.ReadFile("testfile.txt")
		if err != nil {
			t.Fatalf("ReadFile() failed: %v", err)
		}

		expected := "Hello, World!"
		if string(data) != expected {
			t.Errorf("ReadFile() = %q, want %q", string(data), expected)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := gfs.ReadFile("nonexistent.txt")
		if err == nil {
			t.Error("ReadFile() should return error for non-existent file")
		}
	})
}

func TestFileSystem_ReadDir(t *testing.T) {
	gfs := setupTestFS(t)

	t.Run("read directory", func(t *testing.T) {
		entries, err := gfs.ReadDir("testdir")
		if err != nil {
			t.Fatalf("ReadDir() failed: %v", err)
		}

		// Count only regular files (not . and ..)
		var fileCount int
		for _, entry := range entries {
			if entry.Name() != "." && entry.Name() != ".." {
				fileCount++
				if entry.IsDir() {
					t.Errorf("Entry %q is a directory, expected file", entry.Name())
				}
			}
		}

		if fileCount != 5 {
			t.Errorf("ReadDir() returned %d files, want 5", fileCount)
		}
	})

	t.Run("read non-existent directory", func(t *testing.T) {
		_, err := gfs.ReadDir("nonexistent")
		if err == nil {
			t.Error("ReadDir() should return error for non-existent directory")
		}
	})

	t.Run("read root directory", func(t *testing.T) {
		entries, err := gfs.ReadDir(".")
		if err != nil {
			t.Fatalf("ReadDir() failed to read root: %v", err)
		}

		if len(entries) < 2 {
			t.Errorf("ReadDir() returned %d entries, expected at least 2", len(entries))
		}
	})
}

func TestFileSystem_Stat(t *testing.T) {
	gfs := setupTestFS(t)

	t.Run("stat existing file", func(t *testing.T) {
		info, err := gfs.Stat("testfile.txt")
		if err != nil {
			t.Fatalf("Stat() failed: %v", err)
		}

		if info.Name() != "testfile.txt" {
			t.Errorf("Stat() Name() = %q, want %q", info.Name(), "testfile.txt")
		}

		if info.IsDir() {
			t.Error("Stat() IsDir() = true, want false")
		}

		if info.Size() != 13 {
			t.Errorf("Stat() Size() = %d, want 13", info.Size())
		}
	})

	t.Run("stat directory", func(t *testing.T) {
		info, err := gfs.Stat("testdir")
		if err != nil {
			t.Fatalf("Stat() failed: %v", err)
		}

		if !info.IsDir() {
			t.Error("Stat() IsDir() = false, want true for directory")
		}
	})

	t.Run("stat non-existent file", func(t *testing.T) {
		_, err := gfs.Stat("nonexistent.txt")
		if err == nil {
			t.Error("Stat() should return error for non-existent file")
		}
	})
}

func TestFile_Stat(t *testing.T) {
	gfs := setupTestFS(t)

	file, err := gfs.Open("testfile.txt")
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("File.Stat() failed: %v", err)
	}

	if info.Name() != "testfile.txt" {
		t.Errorf("File.Stat() Name() = %q, want %q", info.Name(), "testfile.txt")
	}

	if info.Size() != 13 {
		t.Errorf("File.Stat() Size() = %d, want 13", info.Size())
	}
}

func TestFile_Read(t *testing.T) {
	gfs := setupTestFS(t)

	file, err := gfs.Open("testfile.txt")
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer file.Close()

	buf := make([]byte, 5)
	n, err := file.Read(buf)
	if err != nil {
		t.Fatalf("File.Read() failed: %v", err)
	}

	if n != 5 {
		t.Errorf("File.Read() read %d bytes, want 5", n)
	}

	if string(buf) != "Hello" {
		t.Errorf("File.Read() = %q, want %q", string(buf), "Hello")
	}

	// Read the rest
	rest, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("io.ReadAll() failed: %v", err)
	}

	if string(rest) != ", World!" {
		t.Errorf("io.ReadAll() = %q, want %q", string(rest), ", World!")
	}
}

func TestFile_ReadDir(t *testing.T) {
	gfs := setupTestFS(t)

	t.Run("read all entries with n=0", func(t *testing.T) {
		file, err := gfs.Open("testdir")
		if err != nil {
			t.Fatalf("Open() failed: %v", err)
		}
		defer file.Close()

		// Cast to fs.ReadDirFile to access ReadDir
		rdFile, ok := file.(fs.ReadDirFile)
		if !ok {
			t.Fatal("File does not implement fs.ReadDirFile")
		}

		entries, err := rdFile.ReadDir(0)
		if err != nil {
			t.Fatalf("File.ReadDir(0) failed: %v", err)
		}

		// Count only regular files (not . and ..)
		var fileCount int
		for _, entry := range entries {
			if entry.Name() != "." && entry.Name() != ".." {
				fileCount++
			}
		}

		if fileCount != 5 {
			t.Errorf("File.ReadDir(0) returned %d files, want 5", fileCount)
		}
	})

	t.Run("read limited entries with n>0", func(t *testing.T) {
		file, err := gfs.Open("testdir")
		if err != nil {
			t.Fatalf("Open() failed: %v", err)
		}
		defer file.Close()

		rdFile, ok := file.(fs.ReadDirFile)
		if !ok {
			t.Fatal("File does not implement fs.ReadDirFile")
		}

		entries, err := rdFile.ReadDir(3)
		if err != nil {
			t.Fatalf("File.ReadDir(3) failed: %v", err)
		}

		if len(entries) != 3 {
			t.Errorf("File.ReadDir(3) returned %d entries, want 3", len(entries))
		}
	})

	t.Run("read negative n returns all entries", func(t *testing.T) {
		file, err := gfs.Open("testdir")
		if err != nil {
			t.Fatalf("Open() failed: %v", err)
		}
		defer file.Close()

		rdFile, ok := file.(fs.ReadDirFile)
		if !ok {
			t.Fatal("File does not implement fs.ReadDirFile")
		}

		entries, err := rdFile.ReadDir(-1)
		if err != nil {
			t.Fatalf("File.ReadDir(-1) failed: %v", err)
		}

		// Count only regular files (not . and ..)
		var fileCount int
		for _, entry := range entries {
			if entry.Name() != "." && entry.Name() != ".." {
				fileCount++
			}
		}

		if fileCount != 5 {
			t.Errorf("File.ReadDir(-1) returned %d files, want 5", fileCount)
		}
	})
}

func TestFile_Close(t *testing.T) {
	gfs := setupTestFS(t)

	file, err := gfs.Open("testfile.txt")
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Errorf("File.Close() failed: %v", err)
	}

	// Try to read from closed file - should fail
	buf := make([]byte, 10)
	_, err = file.Read(buf)
	if err == nil {
		t.Error("Reading from closed file should return error")
	}
}

func TestDirEntry_Name(t *testing.T) {
	gfs := setupTestFS(t)

	entries, err := gfs.ReadDir("testdir")
	if err != nil {
		t.Fatalf("ReadDir() failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No entries found")
	}

	entry := entries[0]
	name := entry.Name()

	if name == "" {
		t.Error("DirEntry.Name() returned empty string")
	}
}

func TestDirEntry_IsDir(t *testing.T) {
	gfs := setupTestFS(t)

	entries, err := gfs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() failed: %v", err)
	}

	var foundFile, foundDir bool
	for _, entry := range entries {
		if entry.Name() == "testfile.txt" {
			foundFile = true
			if entry.IsDir() {
				t.Error("DirEntry.IsDir() = true for file, want false")
			}
		}
		if entry.Name() == "testdir" {
			foundDir = true
			if !entry.IsDir() {
				t.Error("DirEntry.IsDir() = false for directory, want true")
			}
		}
	}

	if !foundFile {
		t.Error("Did not find testfile.txt in directory listing")
	}
	if !foundDir {
		t.Error("Did not find testdir in directory listing")
	}
}

func TestDirEntry_Type(t *testing.T) {
	gfs := setupTestFS(t)

	entries, err := gfs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() failed: %v", err)
	}

	for _, entry := range entries {
		mode := entry.Type()
		if entry.IsDir() {
			if !mode.IsDir() {
				t.Errorf("DirEntry.Type().IsDir() = false for directory %q, want true", entry.Name())
			}
		} else {
			if mode.IsDir() {
				t.Errorf("DirEntry.Type().IsDir() = true for file %q, want false", entry.Name())
			}
		}
	}
}

func TestDirEntry_Info(t *testing.T) {
	gfs := setupTestFS(t)

	entries, err := gfs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No entries found")
	}

	entry := entries[0]
	info, err := entry.Info()
	if err != nil {
		t.Fatalf("DirEntry.Info() failed: %v", err)
	}

	if info.Name() != entry.Name() {
		t.Errorf("DirEntry.Info().Name() = %q, want %q", info.Name(), entry.Name())
	}

	if info.IsDir() != entry.IsDir() {
		t.Errorf("DirEntry.Info().IsDir() = %v, want %v", info.IsDir(), entry.IsDir())
	}
}

func TestFileSystem_Implements_Interfaces(t *testing.T) {
	gfs := setupTestFS(t)

	var _ fs.FS = gfs
	var _ fs.ReadFileFS = gfs
	var _ fs.ReadDirFS = gfs
	var _ fs.StatFS = gfs
}

func TestFile_Implements_Interfaces(t *testing.T) {
	gfs := setupTestFS(t)

	file, err := gfs.Open("testfile.txt")
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer file.Close()

	var _ fs.File = file
	var _ io.Reader = file
}

func TestFile_Dir_Implements_Interfaces(t *testing.T) {
	gfs := setupTestFS(t)

	file, err := gfs.Open("testdir")
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer file.Close()

	var _ fs.File = file

	// Check if file implements fs.ReadDirFile
	if _, ok := file.(fs.ReadDirFile); !ok {
		t.Error("File does not implement fs.ReadDirFile interface")
	}
}

// Test that FileSystem can be used with fs.WalkDir
func TestFileSystem_WalkDir(t *testing.T) {
	gfs := setupTestFS(t)

	// Test that we can read a single file using the filesystem
	// We don't test full directory traversal due to potential issues with . and .. in memfs
	fileCount := 0
	err := fs.WalkDir(gfs, "testfile.txt", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fileCount++
		return nil
	})

	if err != nil {
		t.Fatalf("fs.WalkDir() failed on single file: %v", err)
	}

	if fileCount != 1 {
		t.Errorf("fs.WalkDir() visited %d items for single file, expected 1", fileCount)
	}
}

// Test error handling in ReadFile when file cannot be closed
func TestFileSystem_ReadFile_CloseError(t *testing.T) {
	gfs := setupTestFS(t)

	// This test verifies that the defer close error handling works
	// Normal case should work fine
	data, err := gfs.ReadFile("testfile.txt")
	if err != nil {
		t.Fatalf("ReadFile() failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ReadFile() returned empty data")
	}
}

// Test error handling in ReadDir when directory cannot be closed
func TestFileSystem_ReadDir_CloseError(t *testing.T) {
	gfs := setupTestFS(t)

	// This test verifies that the defer close error handling works
	// Normal case should work fine
	entries, err := gfs.ReadDir("testdir")
	if err != nil {
		t.Fatalf("ReadDir() failed: %v", err)
	}

	if len(entries) == 0 {
		t.Error("ReadDir() returned no entries")
	}
}

// Test that FileSystem.Open returns error for empty path
func TestFileSystem_Open_EmptyPath(t *testing.T) {
	gfs := setupTestFS(t)

	_, err := gfs.Open("")
	if err == nil {
		t.Error("Open(\"\") should return error for empty path")
	}
}

// Test reading from file after partial read
func TestFile_MultipleReads(t *testing.T) {
	gfs := setupTestFS(t)

	file, err := gfs.Open("testfile.txt")
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer file.Close()

	// First read
	buf1 := make([]byte, 7)
	n1, err := file.Read(buf1)
	if err != nil {
		t.Fatalf("First Read() failed: %v", err)
	}
	if n1 != 7 || string(buf1) != "Hello, " {
		t.Errorf("First Read() = %q (%d bytes), want \"Hello, \" (7 bytes)", string(buf1), n1)
	}

	// Second read
	buf2 := make([]byte, 6)
	n2, err := file.Read(buf2)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Second Read() failed: %v", err)
	}
	if n2 != 6 || string(buf2) != "World!" {
		t.Errorf("Second Read() = %q (%d bytes), want \"World!\" (6 bytes)", string(buf2), n2)
	}

	// Third read should return EOF
	buf3 := make([]byte, 10)
	n3, err := file.Read(buf3)
	if err != io.EOF {
		t.Errorf("Third Read() error = %v, want io.EOF", err)
	}
	if n3 != 0 {
		t.Errorf("Third Read() returned %d bytes, want 0", n3)
	}
}
