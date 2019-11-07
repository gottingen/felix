package felix

import (
	"os"
	"path/filepath"
	"sort"
)

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
// adapted from https://golang.org/src/path/filepath/path.go
func readDirNames(fs Fs, dirname string) ([]string, error) {
	f, err := fs.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

// walk recursively descends path, calling walkFn
// adapted from https://golang.org/src/path/filepath/path.go
func walk(fs Fs, path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readDirNames(fs, path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := lstatIfPossible(fs, filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(fs, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

// if the filesystem supports it, use Lstat, else use fs.Stat
func lstatIfPossible(fs Fs, path string) (os.FileInfo, error) {
	if lfs, ok := fs.(Lstater); ok {
		fi, _, err := lfs.LstatIfPossible(path)
		return fi, err
	}
	return fs.Stat(path)
}

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.

func (a Felix) Walk(root string, walkFn filepath.WalkFunc) error {
	return Walk(a.Fs, root, walkFn)
}

func Walk(fs Fs, root string, walkFn filepath.WalkFunc) error {
	info, err := lstatIfPossible(fs, root)
	if err != nil {
		return walkFn(root, nil, err)
	}
	return walk(fs, root, info, walkFn)
}
