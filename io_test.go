//go:build go1.12
// +build go1.12

package pty

import (
	"context"
	"errors"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	errMarker byte = 0xEE
	timeout        = time.Second
)

var glTestFdLock sync.Mutex

// Check that SetDeadline() works for ptmx.
// Outstanding Read() calls must be interrupted by deadline.
//
// https://github.com/creack/pty/issues/162
func TestReadDeadline(t *testing.T) {
	ptmx, success := prepare(t)

	if err := ptmx.SetDeadline(time.Now().Add(timeout / 10)); err != nil {
		if errors.Is(err, os.ErrNoDeadline) {
			t.Skipf("Deadline is not supported on %s/%s.", runtime.GOOS, runtime.GOARCH)
		} else {
			t.Fatalf("Error: set deadline: %s.", err)
		}
	}

	buf := make([]byte, 1)
	n, err := ptmx.Read(buf)
	success()
	if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("Unexpected read error: %s.", err)
	}

	if n != 0 && buf[0] != errMarker {
		t.Errorf("Received unexpected data from pmtx (%d bytes): 0x%X; err=%v.", n, buf, err)
	}
}

// Check that ptmx.Close() interrupts outstanding ptmx.Read() calls.
//
// https://github.com/creack/pty/issues/114
// https://github.com/creack/pty/issues/88
func TestReadClose(t *testing.T) {
	ptmx, success := prepare(t)

	go func() {
		time.Sleep(timeout / 10)
		if err := ptmx.Close(); err != nil {
			t.Errorf("Failed to close ptmx: %s.", err)
		}
	}()

	buf := make([]byte, 1)
	n, err := ptmx.Read(buf)
	success()
	if err != nil && !errors.Is(err, os.ErrClosed) {
		t.Fatalf("Unexpected read error: %s.", err)
	}

	if n != 0 && buf[0] != errMarker {
		t.Errorf("Received unexpected data from pmtx (%d bytes): 0x%X; err=%v.", n, buf, err)
	}
}

// Open pty and setup watchdogs for graceful and not so graceful failure modes
func prepare(t *testing.T) (ptmx *os.File, done func()) {
	t.Helper()

	if runtime.GOOS == "darwin" {
		t.Log("creack/pty uses blocking i/o on darwin intentionally:")
		t.Log("> https://github.com/creack/pty/issues/52")
		t.Log("> https://github.com/creack/pty/pull/53")
		t.Log("> https://github.com/golang/go/issues/22099")
		t.SkipNow()
	}

	// Due to data race potential in (*os.File).Fd()
	// we should never run these two tests in parallel.
	glTestFdLock.Lock()
	t.Cleanup(glTestFdLock.Unlock)

	ptmx, pts, err := Open()
	if err != nil {
		t.Fatalf("Error: open: %s.\n", err)
	}
	t.Cleanup(func() { _ = ptmx.Close() })
	t.Cleanup(func() { _ = pts.Close() })

	ctx, done := context.WithCancel(context.Background())
	t.Cleanup(done)
	go func() {
		select {
		case <-ctx.Done():
			// ptmx.Read() did not block forever, yay!
		case <-time.After(timeout):
			if _, err := pts.Write([]byte{errMarker}); err != nil { // unblock ptmx.Read()
				t.Errorf("Failed to write to pts: %s.", err)
			}
			t.Error("ptmx.Read() was not unblocked.")
			done() // cancel panic()
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			// Test has either failed or succeeded; it definitely did not hang.
		case <-time.After(timeout * 10 / 9): // timeout +11%
			panic("ptmx.Read() was not unblocked; avoid hanging forever.") // Just in case.
		}
	}()

	return ptmx, done
}
