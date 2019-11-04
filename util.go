package felix

import (
	"github.com/gottingen/felix/filesystem"
	"io"
	"os"
	"path/filepath"
)


func (a Felix) Walk(root string, walkFn filepath.WalkFunc) error {
	return filesystem.Walk(a.Fs, root, walkFn)
}

func (a Felix) TempDir(dir, prefix string) (name string, err error) {
	return filesystem.TempDir(a.Fs, dir, prefix)
}

func (a Felix) TempFile(dir, prefix string) (f filesystem.File, err error) {
	return filesystem.TempFile(a.Fs, dir, prefix)
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func (a Felix) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return filesystem.WriteFile(a.Fs, filename, data, perm)
}

// ReadFile reads the file named by filename and returns the contents.
// A successful call returns err == nil, not err == EOF. Because ReadFile
// reads the whole file, it does not treat an EOF from Read as an error
// to be reported.
func (a Felix) ReadFile(filename string) ([]byte, error) {
	return filesystem.ReadFile(a.Fs, filename)
}

// ReadDir reads the directory named by dirname and returns
// a list of sorted directory entries.
func (a Felix) ReadDir(dirname string) ([]os.FileInfo, error) {
	return filesystem.ReadDir(a.Fs, dirname)
}


// Same as WriteReader but checks to see if file/directory already exists.
func (a Felix) SafeWriteReader(path string, r io.Reader) (err error) {
	return filesystem.SafeWriteReader(a.Fs, path, r)
}

// Takes a reader and a path and writes the content
func (a Felix) WriteReader(path string, r io.Reader) (err error) {
	return filesystem.WriteReader(a.Fs, path, r)
}


func (a Felix) GetTempDir(subPath string) string {
	return filesystem.GetTempDir(a.Fs, subPath)
}

func (a Felix) FileContainsBytes(filename string, subslice []byte) (bool, error) {
	return filesystem.FileContainsBytes(a.Fs, filename, subslice)
}

func (a Felix) FileContainsAnyBytes(filename string, subslices [][]byte) (bool, error) {
	return filesystem.FileContainsAnyBytes(a.Fs, filename, subslices)
}


func (a Felix) DirExists(path string) (bool, error) {
	return filesystem.DirExists(a.Fs, path)
}

func (a Felix) IsDir(path string) (bool, error) {
	return filesystem.IsDir(a.Fs, path)
}

func (a Felix) IsEmpty(path string) (bool, error) {
	return filesystem.IsEmpty(a.Fs, path)
}

func (a Felix) Exists(path string) (bool, error) {
	return filesystem.Exists(a.Fs, path)
}


