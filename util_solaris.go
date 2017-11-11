package pty

import (
	"os"
	"golang.org/x/sys/unix"
)

const (
	TIOCGWINSZ = 21608
)

// Get Windows Size
func Getsize(t *os.File) (rows, cols int, err error) {
	var wsz *unix.Winsize
	wsz, err = unix.IoctlGetWinsize(int(t.Fd()), TIOCGWINSZ)

	if err != nil {
		return 80, 25, err
	} else {
		return int(wsz.Row), int(wsz.Col), err
	}
}
