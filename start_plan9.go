package pty

import (
	"os"
	"os/exec"
)

// StartWithSize can be supported after a fashion on Plan 9, later.
func StartWithSize(cmd *exec.Cmd, ws *Winsize) (*os.File, error) {
	return nil, ErrUnsupported
}
