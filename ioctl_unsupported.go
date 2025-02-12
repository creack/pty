//go:build aix || plan9
// +build aix plan9

package pty

const (
	TIOCGWINSZ = 0
	TIOCSWINSZ = 0
)

func ioctlInner(fd, cmd, ptr uintptr) error {
	return ErrUnsupported
}
