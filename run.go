package pty

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

// Start assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
// and c.Stderr, calls c.Start, and returns the File of the tty's
// corresponding pty.
func Start(c *exec.Cmd) (pty *os.File, err error) {
	pty, tty, err := Open()
	if err != nil {
		return nil, err
	}
	defer tty.Close()

	if IsTerminal(pty) == false {
		return nil, ErrNotTerminal
	}

	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty
	c.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}
	err = c.Start()
	if err != nil {
		pty.Close()
		return nil, err
	}
	return pty, err
}

// Start raw acts as Start but put the terminal into raw mode. Returns an
// additional function that should be used to restore the terminal state.
func StartRaw(c *exec.Cmd) (pty *os.File, restore func(), err error) {
	pty, err = Start(c)
	oldState, err := MakeRaw(pty)
	if err != nil {
		return nil, nil, err
	}

	return pty, func() {
		var err error
		if err = Restore(pty, oldState); err != nil {
			log.Fatal(err)
		}
	}, nil
}
