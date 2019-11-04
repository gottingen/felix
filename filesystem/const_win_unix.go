
// +build !darwin
// +build !openbsd
// +build !freebsd
// +build !dragonfly
// +build !netbsd

package filesystem

import (
	"syscall"
)

const BADFD = syscall.EBADFD
