package pty

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Setsize resizes t to s.
func Setsize(t FdHolder, ws *Winsize) error {
	resizePseudoConsole, err := kernel32DLL.FindProc("ResizePseudoConsole")
	if err != nil {
		return os.NewSyscallError("ResizePseudoConsole", err)
	}

	_, _, err = resizePseudoConsole.Call(
		t.Fd(),
		uintptr(unsafe.Pointer(&windows.Coord{X: int16(ws.Cols), Y: int16(ws.Rows)})),
	)
	return err
}

// GetsizeFull returns the full terminal size description.
func GetsizeFull(t FdHolder) (size *Winsize, err error) {
	getConsoleScreenBufferInfo, err := kernel32DLL.FindProc("GetConsoleScreenBufferInfo")
	if err != nil {
		return nil, os.NewSyscallError("GetConsoleScreenBufferInfo", err)
	}

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
