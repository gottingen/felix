package vfs

import (
	"os"
)

// Lstater is an optional interface in Felix. It is only implemented by the
// filesystems saying so.
// It will call Lstat if the filesystem iself is, or it delegates to, the os filesystem.
// Else it will call Stat.
// In addtion to the FileInfo, it will return a boolean telling whether Lstat was called or not.
type Lstater interface {
	LstatIfPossible(name string) (os.FileInfo, bool, error)
}