package pty

import (
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	openClose(t)
}

func TestName(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	// Check name isn't empty. There's variation on what exactly the OS calls these files.
	if pty.Name() == "" {
		t.Error("Pty name was empty.")
	}
	if tty.Name() == "" {
		t.Error("Tty name was empty.")
	}
}

// TestOpenByName ensures that the name associated with the tty is valid
// and can be opened and used if passed by file name (rather than passing
// the existing open file descriptor).
func TestOpenByName(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t) // Get the pty/tty pair.

	// Manually open the tty from the exiting name.
	ttyFile, err := os.OpenFile(tty.Name(), os.O_RDWR, 0o600)
	noError(t, err, "Failed to open tty file")
	defer func() { _ = ttyFile.Close() }()

	// Ensure we can write to the newly opened tty file and read on the pty.
	text := []byte("ping")

	n, err := ttyFile.Write(text)                                   // Write the text to the manually open tty.
	noError(t, err, "Unexpected error from Write")                  // Make sure it didn't fail.
	assert(t, len(text), n, "Unexpected count returned from Write") // Assert the number of bytes written.

	buffer := readN(t, pty, len(text), "Unexpected error from pty Read")
	assertBytes(t, text, buffer, "Unexpected result result returned from pty Read")
}

func TestGetsize(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	prows, pcols, err := Getsize(pty)
	noError(t, err, "Unexpected error from pty Getsize")

	trows, tcols, err := Getsize(tty)
	noError(t, err, "Unexpected error from tty Getsize")

	assert(t, prows, trows, "rows from Getsize on pty and tty should match")
	assert(t, pcols, tcols, "cols from Getsize on pty and tty should match")
}

func TestGetsizeFull(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	psize, err := GetsizeFull(pty)
	noError(t, err, "Unexpected error from pty GetsizeFull")

	tsize, err := GetsizeFull(tty)
	noError(t, err, "Unexpected error from tty GetsizeFull")

	assert(t, psize.X, tsize.X, "X from GetsizeFull on pty and tty should match")
	assert(t, psize.Y, tsize.Y, "Y from GetsizeFull on pty and tty should match")
	assert(t, psize.Rows, tsize.Rows, "rows from GetsizeFull on pty and tty should match")
	assert(t, psize.Cols, tsize.Cols, "cols from GetsizeFull on pty and tty should match")
}

func TestSetsize(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	psize, err := GetsizeFull(pty)
	noError(t, err, "Unexpected error from pty GetsizeFull")

	psize.X++
	psize.Y++
	psize.Rows++
	psize.Cols++

	noError(t, Setsize(tty, psize), "Unexpected error from Setsize")

	tsize, err := GetsizeFull(tty)
	noError(t, err, "Unexpected error from tty GetsizeFull")

	assert(t, psize.X, tsize.X, "Unexpected Getsize X result after Setsize")
	assert(t, psize.Y, tsize.Y, "Unexpected Getsize Y result after Setsize")
	assert(t, psize.Rows, tsize.Rows, "Unexpected Getsize Rows result after Setsize")
	assert(t, psize.Cols, tsize.Cols, "Unexpected Getsize Cols result after Setsize")
}

func TestReadWriteText(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	// Write to tty, read from pty.
	{
		text := []byte("ping")

		n, err := tty.Write(text)
		noError(t, err, "Unexpected error from tty Write")
		assert(t, n, len(text), "Unexpected count returned from tty Write")

		buffer := readN(t, pty, len(text), "Unexpected error from pty Read")
		assertBytes(t, text, buffer, "Unexpected result returned from pty Read")
	}

	// Write to pty, read from tty.
	// We need to send a \n otherwise this will block in the terminal driver.
	{
		text := []byte("pong\n")

		n, err := pty.Write(text)
		noError(t, err, "Unexpected error from pty Write")
		assert(t, n, len(text), "Unexpected count returned from pty Write")

		// Expect the raw text back when reading from tty.
		buffer := readN(t, tty, len(text), "Unexpected error from tty Read")
		assertBytes(t, text, buffer, "Unexpected result returned from tty Read")

		// Read the echo back from pty. Expect LF to be CRLF.
		expect := []byte("pong\r\n")
		buffer = readN(t, pty, len(expect), "Unexpected error from pty Read")
		assertBytes(t, expect, buffer, "Unexpected result returned from pty Read")
	}
}

func TestReadWriteControls(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	// Write the start of a line to pty.
	n, err := pty.WriteString("pind")                                    // Intentional typo.
	noError(t, err, "Unexpected error from Write initial text")          // Make sure it didn't fail.
	assert(t, 4, n, "Unexpected count returned from Write initial text") // Assert the number of bytes written.

	// Backspace that last char.
	n, err = pty.WriteString("\b")                                    // "Remove" the typo.
	noError(t, err, "Unexpected error from Write backspace")          // Make sure it didn't fail.
	assert(t, 1, n, "Unexpected count returned from Write backspace") // Assert the number of bytes written.

	// Write the correct char and a LF.
	n, err = pty.WriteString("g\n")                                    // Fix the typo.
	noError(t, err, "Unexpected error from Write fixed text")          // Make sure it didn't fail.
	assert(t, 2, n, "Unexpected count returned from Write fixed text") // Assert the number of bytes written.

	// Read the line.
	buffer := readN(t, tty, 7, "Unexpected error from tty Read")
	assertBytes(t, []byte("pind\bg\n"), buffer, "Unexpected result returned from tty Read")

	// Read the echo back from pty.
	buffer = readN(t, pty, 9, "Unexpected error from pty Read")
	assertBytes(t, []byte("pind^Hg\r\n"), buffer, "Unexpected result returned from pty Read")
}

// Make sure we can have multiple open ptys without mixing them up.
func TestDistinctReadWriteText(t *testing.T) {
	t.Parallel()

	ptyA, ttyA := openClose(t)
	ptyB, ttyB := openClose(t)

	textA := []byte("ping")
	textB := []byte("fizzz")

	n, err := ttyA.Write(textA)
	noError(t, err, "Unexpected error from tty Write A")
	assert(t, n, len(textA), "Unexpected count returned from tty Write A")

	n, err = ttyB.Write(textB)
	noError(t, err, "Unexpected error from tty Write B")
	assert(t, n, len(textB), "Unexpected count returned from tty Write B")

	buffer := readN(t, ptyA, len(textA), "Unexpected error from pty Read A")
	assertBytes(t, textA, buffer, "Unexpected result returned from pty Read B")

	buffer = readN(t, ptyB, len(textB), "Unexpected error from pty Read B")
	assertBytes(t, textB, buffer, "Unexpected result returned from pty Read B")
}

// Make sure we can resize individual ptys without side effect on others.
func TestDistinctSetsize(t *testing.T) {
	t.Parallel()

	_, ttyA := openClose(t)
	_, ttyB := openClose(t)

	noError(t, Setsize(ttyA, &Winsize{Rows: 20}), "Unexpected error from Setsize A")
	noError(t, Setsize(ttyB, &Winsize{Rows: 40}), "Unexpected error from Setsize B")

	tsizeA, err := GetsizeFull(ttyA)
	noError(t, err, "Unexpected error from tty GetsizeFull A")
	tsizeB, err := GetsizeFull(ttyB)
	noError(t, err, "Unexpected error from tty GetsizeFull B")

	assert(t, 20, tsizeA.Rows, "Unexpected Getsize Rows result after Setsize A")
	assert(t, 40, tsizeB.Rows, "Unexpected Getsize Rows result after Setsize B")
}
