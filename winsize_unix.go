//go:build !windows
//+build !windows

package pty

import (
	"syscall"
	"unsafe"
)

// Setsize resizes t to s.
func Setsize(t FdHolder, ws *Winsize) error {
	//nolint:gosec // Expected unsafe pointer for Syscall call.
	return ioctl(t.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(ws)))
}

// GetsizeFull returns the full terminal size description.
func GetsizeFull(t FdHolder) (size *Winsize, err error) {
	var ws Winsize

	//nolint:gosec // Expected unsafe pointer for Syscall call.
	if err := ioctl(t.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws))); err != nil {
		return nil, err
	}
	return &ws, nil
}

// Getsize returns the number of rows (lines) and cols (positions
// in each line) in terminal t.
func Getsize(t FdHolder) (rows, cols int, err error) {
	ws, err := GetsizeFull(t)
	return int(ws.Rows), int(ws.Cols), err
}
