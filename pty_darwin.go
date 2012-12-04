package pty

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	sys_TIOCPTYUNLK  = 0x20007452
	sys_TIOCPTYGRANT = 0x20007454
	sys_TIOCPTYGNAME = 0x40807453
)

// Opens a pty and its corresponding tty.
func Open() (pty, tty *os.File, err error) {
	p, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}

	sname, err := ptsname(p)
	if err != nil {
		return nil, nil, err
	}

	err = grantpt(p)
	if err != nil {
		return nil, nil, err
	}

	err = unlockpt(p)
	if err != nil {
		return nil, nil, err
	}

	t, err := os.OpenFile(sname, os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	return p, t, nil
}

func ptsname(f *os.File) (string, error) {
	var n [64]byte

	ioctl(f.Fd(), sys_TIOCPTYGNAME, uintptr(unsafe.Pointer(&n[0])))

	return string(n[:]), nil
}

func grantpt(f *os.File) error {
	var u int
	return ioctl(f.Fd(), sys_TIOCPTYGRANT, uintptr(unsafe.Pointer(&u)))
}

func unlockpt(f *os.File) error {
	var u int
	return ioctl(f.Fd(), sys_TIOCPTYUNLK, uintptr(unsafe.Pointer(&u)))
}

func ioctl(fd uintptr, cmd uintptr, ptr uintptr) error {
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		cmd,
		uintptr(unsafe.Pointer(ptr)),
	)
	if e != 0 {
		return syscall.ENOTTY
	}
	return nil
}
