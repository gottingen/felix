
// +build !darwin
// +build !openbsd
// +build !freebsd
// +build !dragonfly
// +build !netbsd

package vfs

import (
	"syscall"
)

const BADFD = syscall.EBADFD
