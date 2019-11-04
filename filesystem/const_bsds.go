

// +build darwin openbsd freebsd netbsd dragonfly

package filesystem

import (
"syscall"
)

const BADFD = syscall.EBADF
