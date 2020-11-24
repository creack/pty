package pty

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// InheritSize applies the terminal size of pty to tty. This should be run
// in a signal handler for syscall.SIGWINCH to automatically resize the tty when
// the pty receives a window size change notification.
func InheritSize(pty Pty, tty Tty) error {
	size, err := GetsizeFull(pty)
	if err != nil {
		return err
	}
	err = Setsize(tty, size)
	if err != nil {
		return err
	}
	return nil
}

// Setsize resizes t to s.
func Setsize(t FdHolder, ws *Winsize) error {
	_, _, err := resizePseudoConsole.Call(
		t.Fd(),
		uintptr(unsafe.Pointer(&windows.Coord{X: int16(ws.Cols), Y: int16(ws.Rows)})),
	)
	return err
}

// GetsizeFull returns the full terminal size description.
func GetsizeFull(t FdHolder) (size *Winsize, err error) {
	var info windows.ConsoleScreenBufferInfo
	_, _, err = getConsoleScreenBufferInfo.Call(t.Fd(), uintptr(unsafe.Pointer(&info)))
	return &Winsize{
		Rows: uint16(info.Window.Bottom - info.Window.Top + 1),
		Cols: uint16(info.Window.Right - info.Window.Left + 1),
	}, err
}

// Getsize returns the number of rows (lines) and cols (positions
// in each line) in terminal t.
func Getsize(t FdHolder) (rows, cols int, err error) {
	ws, err := GetsizeFull(t)
	return int(ws.Rows), int(ws.Cols), err
}

type Winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels (not supported)
	Y    uint16 // ws_ypixel: Height in pixels (not supported)
}
