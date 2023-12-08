//go:build !windows && !solaris && !aix
// +build !windows,!solaris,!aix

package pty

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// Local syscall const values.
const (
	TIOCGWINSZ = syscall.TIOCGWINSZ
	TIOCSWINSZ = syscall.TIOCSWINSZ
)

func ioctlInner(fd, cmd, ptr uintptr) error {
	_, _, e := unix.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}
