package pty

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// Will fill p from reader r.
func readBytes(r io.Reader, p []byte) error {
	_, err := io.ReadFull(r, p)
	return err
}

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

	pty, tty := openClose(t)

	ttyFile, err := os.OpenFile(tty.Name(), os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("Failed to open tty file: %s.", err)
	}
	defer func() { _ = ttyFile.Close() }()

	// Ensure we can write to the newly opened tty file and read on the pty.
	text := []byte("ping")
	n, err := ttyFile.Write(text)
	if err != nil {
		t.Fatalf("Unexpected error from Write: %s.", err)
	}
	if n != len(text) {
		t.Errorf("Unexpected count returned from Write, got %d expected %d.", n, len(text))
	}

	buffer := make([]byte, len(text))
	if err := readBytes(pty, buffer); err != nil {
		t.Fatalf("Unexpected error from readBytes: %s.", err)
	}
	if !bytes.Equal(text, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v", buffer, text)
	}
}

func TestGetsize(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	prows, pcols, err := Getsize(pty)
	if err != nil {
		t.Errorf("Unexpected error from Getsize: %s.", err)
	}

	trows, tcols, err := Getsize(tty)
	if err != nil {
		t.Errorf("Unexpected error from Getsize: %s.", err)
	}

	if prows != trows {
		t.Errorf("pty rows != tty rows: %d != %d.", prows, trows)
	}
	if prows != trows {
		t.Errorf("pty cols != tty cols: %d != %d.", pcols, tcols)
	}
}

func TestGetsizefull(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	psize, err := GetsizeFull(pty)
	if err != nil {
		t.Fatalf("Unexpected error from GetsizeFull: %s.", err)
	}

	tsize, err := GetsizeFull(tty)
	if err != nil {
		t.Fatalf("Unexpected error from GetsizeFull: %s.", err)
	}

	if psize.X != tsize.X {
		t.Errorf("pty x != tty x: %d != %d.", psize.X, tsize.X)
	}
	if psize.Y != tsize.Y {
		t.Errorf("pty y != tty y: %d != %d.", psize.Y, tsize.Y)
	}
	if psize.Rows != tsize.Rows {
		t.Errorf("pty rows != tty rows: %d != %d.", psize.Rows, tsize.Rows)
	}
	if psize.Cols != tsize.Cols {
		t.Errorf("pty cols != tty cols: %d != %d.", psize.Cols, tsize.Cols)
	}
}

func TestSetsize(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	psize, err := GetsizeFull(pty)
	if err != nil {
		t.Fatalf("Unexpected error from GetsizeFull: %s.", err)
	}

	psize.X = psize.X + 1
	psize.Y = psize.Y + 1
	psize.Rows = psize.Rows + 1
	psize.Cols = psize.Cols + 1

	if err := Setsize(tty, psize); err != nil {
		t.Fatalf("Unexpected error from Setsize: %s", err)
	}

	tsize, err := GetsizeFull(tty)
	if err != nil {
		t.Fatalf("Unexpected error from GetsizeFull: %s.", err)
	}

	if psize.X != tsize.X {
		t.Errorf("pty x != tty x: %d != %d.", psize.X, tsize.X)
	}
	if psize.Y != tsize.Y {
		t.Errorf("pty y != tty y: %d != %d.", psize.Y, tsize.Y)
	}
	if psize.Rows != tsize.Rows {
		t.Errorf("pty rows != tty rows: %d != %d.", psize.Rows, tsize.Rows)
	}
	if psize.Cols != tsize.Cols {
		t.Errorf("pty cols != tty cols: %d != %d.", psize.Cols, tsize.Cols)
	}
}

func TestReadWriteText(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	// Write to tty, read from pty
	text := []byte("ping")
	n, err := tty.Write(text)
	if err != nil {
		t.Fatalf("Unexpected error from Write: %s", err)
	}
	if n != len(text) {
		t.Errorf("Unexpected count returned from Write, got %d expected %d.", n, len(text))
	}

	buffer := make([]byte, 4)
	if err := readBytes(pty, buffer); err != nil {
		t.Fatalf("Unexpected error from readBytes: %s.", err)
	}
	if !bytes.Equal(text, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v.", buffer, text)
	}

	// Write to pty, read from tty.
	// We need to send a \n otherwise this will block in the terminal driver.
	text = []byte("pong\n")
	n, err = pty.Write(text)
	if err != nil {
		t.Fatalf("Unexpected error from Write: %s.", err)
	}
	if n != len(text) {
		t.Errorf("Unexpected count returned from Write, got %d expected %d.", n, len(text))
	}

	buffer = make([]byte, 5)
	err = readBytes(tty, buffer)
	if err != nil {
		t.Fatalf("Unexpected error from readBytes: %s.", err)
	}
	expect := []byte("pong\n")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v.", buffer, expect)
	}

	// Read the echo back from pty
	buffer = make([]byte, 5)
	err = readBytes(pty, buffer)
	if err != nil {
		t.Fatalf("Unexpected error from readBytes: %s.", err)
	}
	expect = []byte("pong\r")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v.", buffer, expect)
	}
}

func TestReadWriteControls(t *testing.T) {
	t.Parallel()

	pty, tty := openClose(t)

	// Write the start of a line to pty.
	text := []byte("pind")
	n, err := pty.Write(text)
	if err != nil {
		t.Fatalf("Unexpected error from Write: %s.", err)
	}
	if n != len(text) {
		t.Errorf("Unexpected count returned from Write, got %d expected %d.", n, len(text))
	}

	// Backspace that last char.
	n, err = pty.Write([]byte("\b"))
	if err != nil {
		t.Fatalf("Unexpected error from Write: %s.", err)
	}
	if n != 1 {
		t.Errorf("Unexpected count returned from Write, got %d expected %d.", n, 1)
	}

	// Write the correct char and a CR.
	n, err = pty.Write([]byte("g\n"))
	if err != nil {
		t.Fatalf("Unexpected error from Write: %s.", err)
	}
	if n != 2 {
		t.Errorf("Unexpected count returned from Write, got %d expected %d.", n, 2)
	}

	// Read the line
	buffer := make([]byte, 7)
	err = readBytes(tty, buffer)
	if err != nil {
		t.Errorf("Unexpected error from readBytes: %s.", err)
	}
	expect := []byte("pind\bg\n")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v.", buffer, expect)
	}

	// Read the echo back from pty
	buffer = make([]byte, 7)
	err = readBytes(pty, buffer)
	if err != nil {
		t.Errorf("Unexpected error from readBytes: %s.", err)
	}
	expect = []byte("pind^Hg")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v.", buffer, expect)
	}
}
