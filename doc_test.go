package pty

import (
	"bytes"
	"io"
	"testing"
)

// Will fill p from reader r
func readBytes(r io.Reader, p []byte) error {
	bytesRead := 0
	for bytesRead < len(p) {
		n, err := r.Read(p[bytesRead:])
		if err != nil {
			return err
		}
		bytesRead = bytesRead + n
	}
	return nil
}

func TestOpen(t *testing.T) {
	t.Parallel()

	pty, tty, err := Open()
	if err != nil {
		t.Errorf("Unexpected error from Open: %s", err)
	}

	err = tty.Close()
	if err != nil {
		t.Errorf("Unexpected error from tty Close: %s", err)
	}

	err = pty.Close()
	if err != nil {
		t.Errorf("Unexpected error from pty Close: %s", err)
	}
}

func TestName(t *testing.T) {
	t.Parallel()

	pty, tty, err := Open()
	if err != nil {
		t.Errorf("Unexpected error from Open: %s", err)
	}

	// Check name isn't empty. There's variation on what exactly the OS calls these files.
	if pty.Name() == "" {
		t.Error("pty name was empty")
	}
	if tty.Name() == "" {
		t.Error("tty name was empty")
	}

	err = tty.Close()
	if err != nil {
		t.Errorf("Unexpected error from tty Close: %s", err)
	}

	err = pty.Close()
	if err != nil {
		t.Errorf("Unexpected error from pty Close: %s", err)
	}
}

func TestGetsize(t *testing.T) {
	t.Parallel()

	pty, tty, err := Open()
	if err != nil {
		t.Errorf("Unexpected error from Open: %s", err)
	}

	prows, pcols, err := Getsize(pty)
	if err != nil {
		t.Errorf("Unexpected error from Getsize: %s", err)
	}

	trows, tcols, err := Getsize(tty)
	if err != nil {
		t.Errorf("Unexpected error from Getsize: %s", err)
	}

	if prows != trows {
		t.Errorf("pty rows != tty rows: %d != %d", prows, trows)
	}
	if prows != trows {
		t.Errorf("pty cols != tty cols: %d != %d", pcols, tcols)
	}

	err = tty.Close()
	if err != nil {
		t.Errorf("Unexpected error from tty Close: %s", err)
	}

	err = pty.Close()
	if err != nil {
		t.Errorf("Unexpected error from pty Close: %s", err)
	}
}

func TestGetsizefull(t *testing.T) {
	t.Parallel()

	pty, tty, err := Open()
	if err != nil {
		t.Errorf("Unexpected error from Open: %s", err)
	}

	psize, err := GetsizeFull(pty)
	if err != nil {
		t.Errorf("Unexpected error from GetsizeFull: %s", err)
	}

	tsize, err := GetsizeFull(tty)
	if err != nil {
		t.Errorf("Unexpected error from GetsizeFull: %s", err)
	}

	if psize.X != tsize.X {
		t.Errorf("pty x != tty x: %d != %d", psize.X, tsize.X)
	}
	if psize.Y != tsize.Y {
		t.Errorf("pty y != tty y: %d != %d", psize.Y, tsize.Y)
	}
	if psize.Rows != tsize.Rows {
		t.Errorf("pty rows != tty rows: %d != %d", psize.Rows, tsize.Rows)
	}
	if psize.Cols != tsize.Cols {
		t.Errorf("pty cols != tty cols: %d != %d", psize.Cols, tsize.Cols)
	}

	err = tty.Close()
	if err != nil {
		t.Errorf("Unexpected error from tty Close: %s", err)
	}

	err = pty.Close()
	if err != nil {
		t.Errorf("Unexpected error from pty Close: %s", err)
	}
}

func TestSetsize(t *testing.T) {
	t.Parallel()

	pty, tty, err := Open()
	if err != nil {
		t.Errorf("Unexpected error from Open: %s", err)
	}

	psize, err := GetsizeFull(pty)
	if err != nil {
		t.Errorf("Unexpected error from GetsizeFull: %s", err)
	}

	psize.X = psize.X + 1
	psize.Y = psize.Y + 1
	psize.Rows = psize.Rows + 1
	psize.Cols = psize.Cols + 1

	err = Setsize(tty, psize)
	if err != nil {
		t.Errorf("Unexpected error from Setsize: %s", err)
	}

	tsize, err := GetsizeFull(tty)
	if err != nil {
		t.Errorf("Unexpected error from GetsizeFull: %s", err)
	}

	if psize.X != tsize.X {
		t.Errorf("pty x != tty x: %d != %d", psize.X, tsize.X)
	}
	if psize.Y != tsize.Y {
		t.Errorf("pty y != tty y: %d != %d", psize.Y, tsize.Y)
	}
	if psize.Rows != tsize.Rows {
		t.Errorf("pty rows != tty rows: %d != %d", psize.Rows, tsize.Rows)
	}
	if psize.Cols != tsize.Cols {
		t.Errorf("pty cols != tty cols: %d != %d", psize.Cols, tsize.Cols)
	}

	err = tty.Close()
	if err != nil {
		t.Errorf("Unexpected error from tty Close: %s", err)
	}

	err = pty.Close()
	if err != nil {
		t.Errorf("Unexpected error from pty Close: %s", err)
	}
}

func TestReadWriteText(t *testing.T) {
	t.Parallel()

	pty, tty, err := Open()
	if err != nil {
		t.Errorf("Unexpected error from Open: %s", err)
	}

	// Write to tty, read from pty
	text := []byte("ping")
	n, err := tty.Write(text)
	if err != nil {
		t.Errorf("Unexpected error from Write: %s", err)
	}
	if n != len(text) {
		t.Errorf("Unexpected count returned from Write, got %d expected %d", n, len(text))
	}

	buffer := make([]byte, 4)
	err = readBytes(pty, buffer)
	if err != nil {
		t.Errorf("Unexpected error from readBytes: %s", err)
	}
	if !bytes.Equal(text, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v", buffer, text)
	}

	// Write to pty, read from tty.
	// We need to send a \n otherwise this will block in the terminal driver.
	text = []byte("pong\n")
	n, err = pty.Write(text)
	if err != nil {
		t.Errorf("Unexpected error from Write: %s", err)
	}
	if n != len(text) {
		t.Errorf("Unexpected count returned from Write, got %d expected %d", n, len(text))
	}

	buffer = make([]byte, 5)
	err = readBytes(tty, buffer)
	if err != nil {
		t.Errorf("Unexpected error from readBytes: %s", err)
	}
	expect := []byte("pong\n")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v", buffer, expect)
	}

	// Read the echo back from pty
	buffer = make([]byte, 5)
	err = readBytes(pty, buffer)
	if err != nil {
		t.Errorf("Unexpected error from readBytes: %s", err)
	}
	expect = []byte("pong\r")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v", buffer, expect)
	}

	err = tty.Close()
	if err != nil {
		t.Errorf("Unexpected error from tty Close: %s", err)
	}

	err = pty.Close()
	if err != nil {
		t.Errorf("Unexpected error from pty Close: %s", err)
	}
}

func TestReadWriteControls(t *testing.T) {
	t.Parallel()

	pty, tty, err := Open()
	if err != nil {
		t.Errorf("Unexpected error from Open: %s", err)
	}

	// Write the start of a line to pty
	text := []byte("pind")
	n, err := pty.Write(text)
	if err != nil {
		t.Errorf("Unexpected error from Write: %s", err)
	}
	if n != len(text) {
		t.Errorf("Unexpected count returned from Write, got %d expected %d", n, len(text))
	}

	// Backspace that last char
	n, err = pty.Write([]byte("\b"))
	if err != nil {
		t.Errorf("Unexpected error from Write: %s", err)
	}
	if n != 1 {
		t.Errorf("Unexpected count returned from Write, got %d expected %d", n, 1)
	}

	// Write the correct char and a CR
	n, err = pty.Write([]byte("g\n"))
	if err != nil {
		t.Errorf("Unexpected error from Write: %s", err)
	}
	if n != 2 {
		t.Errorf("Unexpected count returned from Write, got %d expected %d", n, 2)
	}

	// Read the line
	buffer := make([]byte, 7)
	err = readBytes(tty, buffer)
	if err != nil {
		t.Errorf("Unexpected error from readBytes: %s", err)
	}
	expect := []byte("pind\bg\n")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v", buffer, expect)
	}

	// Read the echo back from pty
	buffer = make([]byte, 7)
	err = readBytes(pty, buffer)
	if err != nil {
		t.Errorf("Unexpected error from readBytes: %s", err)
	}
	expect = []byte("pind^Hg")
	if !bytes.Equal(expect, buffer) {
		t.Errorf("Unexpected result returned from Read, got %v expected %v", buffer, expect)
	}

	err = tty.Close()
	if err != nil {
		t.Errorf("Unexpected error from tty Close: %s", err)
	}

	err = pty.Close()
	if err != nil {
		t.Errorf("Unexpected error from pty Close: %s", err)
	}
}
