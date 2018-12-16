package pty

import (
	"os"
	"syscall"
	"unsafe"
)

func open() (pty, tty *os.File, err error) {
	/*
	 * from ptm(4):
         * ... TIOCPTMGET, which allocates a free pseudo-terminal device,
         * sets its user ID to the calling user, revoke(2)s it, and
         * returns the opened file descriptors for both the master and the slave
         * pseudo-terminal device to the caller in a struct ptmget.
	 */

	p, err := os.OpenFile("/dev/ptm", os.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	defer p.Close()

	var ptm ptmget
	if err := ioctl(p.Fd(), uintptr(ioctl_TIOCPTMGET), uintptr(unsafe.Pointer(&ptm))); err != nil {
		return nil, nil, err
	}

	pty = os.NewFile(uintptr(ptm.Cfd), "/dev/ptm")
	tty = os.NewFile(uintptr(ptm.Sfd), "/dev/ptm")

	return pty, tty, nil
}
