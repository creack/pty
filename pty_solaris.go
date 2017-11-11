package pty

/* based on:
http://src.illumos.org/source/xref/illumos-gate/usr/src/lib/libc/port/gen/pt.c
*/

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

const NODEV = ^uint64(0)

func open() (pty, tty *os.File, err error) {
	pFD, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	p := os.NewFile(uintptr(pFD), "/dev/ptmx")

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

func minor(x uint64) uint64 {
	return x & 0377
}

func ptsdev(fd uintptr) uint64 {
	istr := strioctl{ISPTM, 0, 0, nil}
	err := ioctl(fd, I_STR, uintptr(unsafe.Pointer(&istr)))
	if err != nil {
		return NODEV
	}
	var status unix.Stat_t
	err = unix.Fstat(int(fd), &status)
	if err != nil {
		return NODEV
	}
	return uint64(minor(status.Rdev))
}

func ptsname(f *os.File) (string, error) {
	dev := ptsdev(f.Fd())
	if dev == NODEV {
		return "", errors.New("not a master pty")
	}
	return "/dev/pts/" + strconv.FormatInt(int64(dev), 10), nil
}

type pt_own struct {
	pto_ruid int32
	pto_rgid int32
}

func grantpt(f *os.File) error {
	if ptsdev(f.Fd()) == NODEV {
		return errors.New("not a master pty")
	}
	var pto pt_own
	pto.pto_ruid = int32(os.Getuid())
	// XXX should first attempt to get gid of DEFAULT_TTY_GROUP="tty"
	pto.pto_rgid = int32(os.Getgid())
	var istr strioctl
	istr.ic_cmd = OWNERPT
	istr.ic_timout = 0
	istr.ic_len = int32(unsafe.Sizeof(istr))
	istr.ic_dp = unsafe.Pointer(&pto)
	err := ioctl(f.Fd(), I_STR, uintptr(unsafe.Pointer(&istr)))
	if err != nil {
		return errors.New("access denied")
	}
	return nil
}

func unlockpt(f *os.File) error {
	istr := strioctl{UNLKPT, 0, 0, nil}
	return ioctl(f.Fd(), I_STR, uintptr(unsafe.Pointer(&istr)))
}
