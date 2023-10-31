package pty

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

// openCloses opens a pty/tty pair and stages the closing as part of the cleanup.
func openClose(t *testing.T) (pty, tty *os.File) {
	t.Helper()

	pty, tty, err := Open()
	if err != nil {
		t.Fatalf("Unexpected error from Open: %s.", err)
	}
	t.Cleanup(func() {
		if err := tty.Close(); err != nil {
			t.Errorf("Unexpected error from tty Close: %s.", err)
		}

		if err := pty.Close(); err != nil {
			t.Errorf("Unexpected error from pty Close: %s.", err)
		}
	})

	return pty, tty
}

func noError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %s.", msg, err)
	}
}

func assert[T comparable](t *testing.T, a, b T, msg string) {
	t.Helper()
	if a != b {
		t.Errorf("%s: %v != %v.", msg, a, b)
	}
}

// When asserting bytes, we want to display the diff, which may contain control characters.
func assertBytes(t *testing.T, a, b []byte, msg string) {
	t.Helper()
	if !bytes.Equal(a, b) {
		t.Errorf("%s: %v != %v.", msg, a, b)
	}
}

// Read a specific number of bytes from r, assert it didn't fail and return the bytes.
func readN(t *testing.T, r io.Reader, n int, msg string) []byte {
	t.Helper()

	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	noError(t, err, fmt.Sprintf("%s: %s", msg, err))
	return buf
}
