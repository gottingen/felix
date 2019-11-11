// +build darwin openbsd freebsd netbsd dragonfly

package vfs

import (
	"syscall"
)

const BADFD = syscall.EBADF
