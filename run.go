package pty

import (
	"exec"
	"os"
)


// Run starts a process with its stdin, stdout, and stderr
// connected to a pseudo-terminal tty;
// Stdin and Stdout of the returned exec.Cmd
// are the corresponding pty (Stderr is always nil).
// Arguments name, argv, envv, and dir are passed
// to os.StartProcess unchanged.
func Run(name string, argv, envv []string, dir string) (c *exec.Cmd, err os.Error) {
	c = new(exec.Cmd)
	var fd [3]*os.File

	c.Stdin, fd[0], err = Open()
	if err != nil {
		return nil, err
	}
	fd[1] = fd[0]
	fd[2] = fd[0]
	c.Stdout = c.Stdin
	c.Process, err = os.StartProcess(name, argv, envv, dir, fd[:])
	fd[0].Close()
	if err != nil {
		c.Stdin.Close()
		return nil, err
	}
	return c, nil
}
