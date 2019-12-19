package pty_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"

	"github.com/creack/pty"
)

type testCmd struct {
	*exec.Cmd
	stdoutPipe io.Reader
}

func testDaemon(command string, args ...string) (*testCmd, error) {
	// Create the command.
	cmd := exec.Command(command, args...)

	// Create a pipe to let the parent consume the output directly.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("cmd.StdoutPipe: %w", err)
	}
	cmd.Stderr = cmd.Stdout
	cmd.Env = os.Environ()

	// Daemonize the process.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// Start the process.
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cmd.Start: %w", err)
	}

	return &testCmd{Cmd: cmd, stdoutPipe: stdoutPipe}, nil
}

func testExec() ([]byte, error) {
	_ = os.Setenv("forked_TestDaemonizedPTY", "test_fork_noctty_cmd")
	defer func() { _ = os.Unsetenv("forked_TestDaemonizedPTY") }()

	// Create a command with "grep".
	c := exec.Command(os.Args[0], os.Args[1:]...)
	c.Env = os.Environ()

	// Start the command in a pty.
	f, err := pty.Start(c)
	defer func() { _ = f.Close() }() // Best effort.
	if err != nil {
		return nil, fmt.Errorf("pty.Start: %w", err)
	}

	// Return the output of grep.
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		// Make sure we have the expected ptmx close error and not something else.
		var pErr *os.PathError
		if !errors.As(err, &pErr) || pErr.Op != "read" || pErr.Path != "/dev/ptmx" {
			return nil, err
		}
	}
	return buf, nil
}

func init() {
	// If the binary is ran with that env variable, we are running the daemonized noctty test.
	if forkedEnv := os.Getenv("forked_TestDaemonizedPTY"); forkedEnv == "test_fork_noctty" {
		buf, err := testExec()
		if err != nil {
			log.Fatalf("testCmd: %s", err)
		}
		fmt.Printf("%s", buf)
		os.Exit(0)
	} else if forkedEnv == "test_fork_noctty_cmd" {
		// If we have this env, we are running a "regular" exec test.
		fmt.Printf("OK!")
		os.Exit(0)
	}
}

// Global fork lock to make sure the test runs only once at a time.
var testDaemonizePTYForkLock sync.Mutex

// NOCTTY test:
// Start a daemon which starts a command with a pty.
func TestDaemonizedPTY(t *testing.T) {
	// Lock the test to avoid issues with fork.
	testDaemonizePTYForkLock.Lock()
	defer testDaemonizePTYForkLock.Unlock()

	// Set an env variable to tell the daemonized child to run the proper test.
	_ = os.Setenv("forked_TestDaemonizedPTY", "test_fork_noctty")
	defer func() { _ = os.Unsetenv("forked_TestDaemonizedPTY") }()

	// Re-run the current binary in daemon mode (no ctty).
	cmd, err := testDaemon(os.Args[0], os.Args[1:]...)
	if err != nil {
		t.Fatalf("Fork: %s.\n", err)
	}
	// Consume the output of the child process.
	buf, err := ioutil.ReadAll(cmd.stdoutPipe)
	if err != nil {
		t.Fatalf("Read: %s.\n", err)
	}
	// Assert result.
	if expect, got := "OK!", string(buf); expect != got {
		t.Fatalf("Unexpected result.\nExpect:\t%q\nGot:\t%q\n", expect, got)
	}
}
