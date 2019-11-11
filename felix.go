package felix

import (
	"github.com/gottingen/felix/vfs"
	"io"
	"os"
	"path/filepath"
)

type Felix struct {
	Vfs vfs.Vfs
}


func NewOsVfs() Felix {
	return Felix{vfs.NewOsFs()}
}

func NewMemVfs() Felix {
	return Felix{vfs.NewMemMapFs()}
}

func NewReadOnlyFs(felix Felix) Felix {
	return Felix{vfs.NewReadOnlyFs(felix.Vfs)}
}

func NewCopyOnWriteFs(base Felix, layer Felix) Felix {
	return Felix{vfs.NewCopyOnWriteFs(base.Vfs, layer.Vfs)}
}


func (a Felix) TempFile(dir, prefix string) (f vfs.File, err error) {
	return vfs.TempFile(a.Vfs, dir, prefix)
}

func (a Felix) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return vfs.WriteFile(a.Vfs, filename, data, perm)
}

func (a Felix) ReadFile(filename string) ([]byte, error) {
	return vfs.ReadFile(a.Vfs, filename)
}

func (a Felix) ReadDir(dirname string) ([]os.FileInfo, error) {
	return vfs.ReadDir(a.Vfs, dirname)
}

// Takes a reader and a path and writes the content
func (a Felix) WriteReader(path string, r io.Reader) (err error) {
	return vfs.WriteReader(a.Vfs, path, r)
}

// Same as WriteReader but checks to see if file/directory already exists.
func (a Felix) SafeWriteReader(path string, r io.Reader) (err error) {
	return vfs.SafeWriteReader(a.Vfs, path, r)
}

func (a Felix) GetTempDir(subPath string) string {
	return vfs.GetTempDir(a.Vfs, subPath)
}

func (a Felix) FileContainsBytes(filename string, subslice []byte) (bool, error) {
	return vfs.FileContainsBytes(a.Vfs, filename, subslice)
}

func (a Felix) DirExists(path string) (bool, error) {
	return vfs.DirExists(a.Vfs, path)
}

func (a Felix) IsDir(path string) (bool, error) {
	return vfs.IsDir(a.Vfs, path)
}

func (a Felix) IsEmpty(path string) (bool, error) {
	return vfs.IsEmpty(a.Vfs, path)
}

func (a Felix) Exists(path string) (bool, error) {
	return vfs.Exists(a.Vfs, path)
}

func (a Felix) Walk(root string, walkFn filepath.WalkFunc) error {
	return vfs.Walk(a.Vfs, root, walkFn)
}