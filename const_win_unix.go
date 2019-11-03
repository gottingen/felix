
// +build !darwin
// +build !openbsd
// +build !freebsd
// +build !dragonfly
// +build !netbsd

package felix

import (
	"syscall"
)

const BADFD = syscall.EBADFD
