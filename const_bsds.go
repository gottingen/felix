

// +build darwin openbsd freebsd netbsd dragonfly

package felix

import (
"syscall"
)

const BADFD = syscall.EBADF
