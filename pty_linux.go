package pty

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	sys_TIOCGPTN   = 0x80045430
	sys_TIOCSPTLCK = 0x40045431
)


// Opens a pty and its corresponding tty.
func Open() (pty, tty *os.File, err os.Error) {
	p, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}

	sname, err := ptsname(p)
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


func ptsname(f *os.File) (string, os.Error) {
	var n int
	err := ioctl(f.Fd(), sys_TIOCGPTN, &n)
	if err != nil {
		return "", err
	}
	return "/dev/pts/" + strconv.Itoa(n), nil
}


func unlockpt(f *os.File) os.Error {
	var u int
	return ioctl(f.Fd(), sys_TIOCSPTLCK, &u)
}


func ioctl(fd int, cmd uint, data *int) os.Error {
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(cmd),
		uintptr(unsafe.Pointer(data)),
	)
	if e != 0 {
		return os.ENOTTY
	}
	return nil
}
