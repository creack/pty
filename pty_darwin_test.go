//go:build darwin
// +build darwin

package pty

import (
	"bufio"
	"testing"
)

func TestDarwinIUTF8(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	// Write multibyte UTF-8 characters followed by Delete (0x7f) and a newline.
	// With IUTF8 enabled, erasing deletes the entire multibyte character rather than only the final byte.
	const input = "小明\x7f\n"
	n, err := pty.WriteString(input)
	noError(t, err, "Unexpected error writing input")
	assert(t, len(input), n, "Unexpected count returned from WriteString")

	r := bufio.NewReader(tty)
	line, err := r.ReadString('\n')
	noError(t, err, "Unexpected error reading from tty")

	expected := "小\n"
	assert(t, expected, line, "Expected multibyte character to be deleted cleanly")
}
