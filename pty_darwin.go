package pty

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	sys_TIOCGPTN   = 0x80045430
	sys_TIOCSPTLCK = 0x40045431
)


// Opens a pty and its corresponding tty.
func Open() (pty, tty *os.File, err os.Error) {
	p, err := os.Open("/dev/ptmx", os.O_RDWR, 0)
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

	t, err := os.Open(sname, os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	return p, t, nil
}

const (
	ptdev1 = "pqrsPQRS"
	ptdev2 = "0123456789abcdefghijklmnopqrstuv"
)

func ptsname(f *os.File) (string, os.Error) {
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	return "/dev/tty" + string([]byte{
		ptdev1[minor(fi.Rdev)/32],
		ptdev2[minor(fi.Rdev)%32],
	}), nil
}


func grantpt(f *os.File) os.Error {
	p, err := os.StartProcess("/bin/ptchown", []string{"/bin/ptchown"},
nil, "", []*os.File{f})
	if err != nil {
		return err
	}
	w, err := p.Wait(0)
	if err != nil {
		return err
	}
	if w.Exited() && w.ExitStatus() == 0 {
		return nil
	}
	return os.EACCES
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


func minor(d uint64) int {
	return int(d & 0xffffffff)
}
