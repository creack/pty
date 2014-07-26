// Package pty provides functions for working with Unix terminals.
package pty

import (
	"errors"
	"os"
)

// ErrUnsupported is returned if a function is not
// available on the current platform.
var ErrUnsupported = errors.New("unsupported")

// ErrNotTerminal is returned if the file descriptor
// picked up is not a tty
var ErrNotTerminal = errors.New("fd is not a terminal")

// Opens a pty and its corresponding tty.
func Open() (pty, tty *os.File, err error) {
	return open()
}
