package pty

import (
	"os"
	"syscall"
	"unsafe"
)

// Getsize returns the number of rows (lines) and cols (positions
// in each line) in terminal t.
func Getsize(t *os.File) (rows, cols int, err error) {
	var ws winsize
	err = windowrect(&ws, t.Fd())
	return int(ws.ws_row), int(ws.ws_col), err
}

type winsize struct {
	ws_row    uint16
	ws_col    uint16
	ws_xpixel uint16
	ws_ypixel uint16
}

func windowrect(ws *winsize, fd uintptr) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(ws)),
	)
	if errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}

// State contains the state of a terminal.
type State struct {
	termios syscall.Termios
}

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(t *os.File) bool {
	var state State
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, t.Fd(), ioctlReadTermios, uintptr(unsafe.Pointer(&state.termios)))
	return err == 0
}

// MakeRaw put the terminal connected to the given file descriptor into raw
// mode and returns the previous state of the terminal so that it can be
// restored.
func MakeRaw(t *os.File) (*State, error) {
	var oldState State
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, t.Fd(), ioctlReadTermios, uintptr(unsafe.Pointer(&oldState.termios))); err != 0 {
		return nil, err
	}

	newState := oldState.termios
	newState.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF
	newState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, t.Fd(), ioctlWriteTermios, uintptr(unsafe.Pointer(&newState))); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

// Restore restores the terminal connected to the given file descriptor to a
// previous state.
func Restore(t *os.File, state *State) error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, t.Fd(), ioctlWriteTermios, uintptr(unsafe.Pointer(&state.termios)))
	return err
}
