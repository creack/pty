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

var mu sync.Mutex

// Check that SetDeadline() works for ptmx.
// Outstanding Read() calls must be interrupted by deadline.
//
// https://github.com/creack/pty/issues/162
func TestReadDeadline(t *testing.T) {
	ptmx, success := prepare(t)

	err := ptmx.SetDeadline(time.Now().Add(timeout / 10))
	if err != nil {
		if errors.Is(err, os.ErrNoDeadline) {
			t.Skipf("deadline is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
		} else {
			t.Fatalf("error: set deadline: %v\n", err)
		}
	}

	var buf = make([]byte, 1)
	n, err := ptmx.Read(buf)
	success()

	if n != 0 && buf[0] != errMarker {
		t.Errorf("received unexpected data from pmtx (%d bytes): 0x%X; err=%v", n, buf, err)
	}
}

// Check that ptmx.Close() interrupts outstanding ptmx.Read() calls
//
// https://github.com/creack/pty/issues/114
// https://github.com/creack/pty/issues/88
func TestReadClose(t *testing.T) {
	ptmx, success := prepare(t)

	go func() {
		time.Sleep(timeout / 10)
		err := ptmx.Close()
		if err != nil {
			t.Errorf("failed to close ptmx: %v", err)
		}
	}()

	var buf = make([]byte, 1)
	n, err := ptmx.Read(buf)
	success()

	if n != 0 && buf[0] != errMarker {
		t.Errorf("received unexpected data from pmtx (%d bytes): 0x%X; err=%v", n, buf, err)
	}
}

// Open pty and setup watchdogs for graceful and not so graceful failure modes
func prepare(t *testing.T) (ptmx Pty, done func()) {
	if runtime.GOOS == "darwin" {
		t.Log("creack/pty uses blocking i/o on darwin intentionally:")
		t.Log("> https://github.com/creack/pty/issues/52")
		t.Log("> https://github.com/creack/pty/pull/53")
		t.Log("> https://github.com/golang/go/issues/22099")
		t.SkipNow()
	}

	// Due to data race potential in (*os.File).Fd()
	// we should never run these two tests in parallel
	mu.Lock()
	t.Cleanup(mu.Unlock)

	ptmx, pts, err := Open()
	if err != nil {
		t.Fatalf("error: open: %v\n", err)
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
			_, err := pts.Write([]byte{errMarker}) // unblock ptmx.Read()
			if err != nil {
				t.Errorf("failed to write to pts: %v", err)
			}
			t.Error("ptmx.Read() was not unblocked")
			done() // cancel panic()
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			// Test has either failed or succeeded; it definitely did not hang
		case <-time.After(timeout * 10 / 9): // timeout +11%
			panic("ptmx.Read() was not unblocked; avoid hanging forever") // just in case
		}
	}()
	return ptmx, done
}
