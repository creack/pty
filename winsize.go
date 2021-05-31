package pty

// Winsize describes the terminal size.
type Winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

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
