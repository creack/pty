package pty

import (
	"os"
	"unsafe"
)

// types from golang.org/x/sys/windows
type (
	// copy of https://pkg.go.dev/golang.org/x/sys/windows#Coord
	windowsCoord struct {
		X int16
		Y int16
	}

	// copy of https://pkg.go.dev/golang.org/x/sys/windows#ConsoleScreenBufferInfo
	windowsConsoleScreenBufferInfo struct {
		Size              Coord
		CursorPosition    Coord
		Attributes        uint16
		Window            SmallRect
		MaximumWindowSize Coord
	}
)

// Setsize resizes t to s.
func Setsize(t FdHolder, ws *Winsize) error {
	resizePseudoConsole, err := kernel32DLL.FindProc("ResizePseudoConsole")
	if err != nil {
		return os.NewSyscallError("ResizePseudoConsole", err)
	}

	_, _, err = resizePseudoConsole.Call(
		t.Fd(),
		uintptr(unsafe.Pointer(&windowsCoord{X: int16(ws.Cols), Y: int16(ws.Rows)})),
	)
	return err
}

// GetsizeFull returns the full terminal size description.
func GetsizeFull(t FdHolder) (size *Winsize, err error) {
	getConsoleScreenBufferInfo, err := kernel32DLL.FindProc("GetConsoleScreenBufferInfo")
	if err != nil {
		return nil, os.NewSyscallError("GetConsoleScreenBufferInfo", err)
	}

	var info windowsConsoleScreenBufferInfo
	_, _, err = getConsoleScreenBufferInfo.Call(t.Fd(), uintptr(unsafe.Pointer(&info)))
	return &Winsize{
		Rows: uint16(info.Window.Bottom - info.Window.Top + 1),
		Cols: uint16(info.Window.Right - info.Window.Left + 1),
	}, err
}
