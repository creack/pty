//go:build aix || plan9
// +build aix plan9

package pty

import "unsafe"

const (
	TIOCGWINSZ = 0
	TIOCSWINSZ = 0
)

func ioctlInner(fd, cmd uintptr, ptr unsafe.Pointer) error {
	return ErrUnsupported
}
