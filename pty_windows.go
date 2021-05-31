//go:build windows
//+build windows

package pty

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32DLL *syscall.DLL
)

func init() {
	var err error
	kernel32DLL, err = syscall.LoadDLL("kernel32.dll")
	if err != nil {
		panic(err)
	}
}

func open() (_ Pty, _ Tty, err error) {
	pr, consoleW, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	consoleR, pw, err := os.Pipe()
	if err != nil {
		_ = consoleW.Close()
		_ = pr.Close()
		return nil, nil, err
	}

	defer func() {
		if err != nil {
			_ = consoleW.Close()
			_ = pr.Close()

			_ = pw.Close()
			_ = consoleR.Close()
		}
	}()

	var (
		consoleHandle syscall.Handle
		defaultSize   = &windowsCoord{X: 80, Y: 30}
	)

	// https://docs.microsoft.com/en-us/windows/console/createpseudoconsole

	createPseudoConsole, err := kernel32DLL.FindProc("CreatePseudoConsole")
	if err != nil {
		return nil, nil, os.NewSyscallError("CreatePseudoConsole", err)
	}

	r1, _, err := createPseudoConsole.Call(
		uintptr(unsafe.Pointer(defaultSize)),    // size: default 80x30 window
		consoleR.Fd(),                           // console input
		consoleW.Fd(),                           // console output
		0,                                       // console flags, currently only PSEUDOCONSOLE_INHERIT_CURSOR supported
		uintptr(unsafe.Pointer(&consoleHandle)), // console handler value return
	)
	if r1 != 0 {
		// S_OK: 0
		return nil, nil, os.NewSyscallError("CreatePseudoConsole", err)
	}

	return &WindowsPty{
			handle:    uintptr(consoleHandle),
			r:         pr,
			w:         pw,
			consoleR:  consoleR,
			consoleW: consoleW,
		}, &WindowsTty{
			handle: uintptr(consoleHandle),
			r:      consoleR,
			w:      consoleW,
		}, nil
}

var _ Pty = (*WindowsPty)(nil)

type WindowsPty struct {
	handle uintptr
	r, w   *os.File

	consoleR, consoleW *os.File
}

func (p *WindowsPty) Fd() uintptr {
	return p.handle
}

func (p *WindowsPty) Read(data []byte) (int, error) {
	return p.r.Read(data)
}

func (p *WindowsPty) Write(data []byte) (int, error) {
	return p.w.Write(data)
}

func (p *WindowsPty) WriteString(s string) (int, error) {
	return p.w.WriteString(s)
}

func (p *WindowsPty) InputPipe() *os.File {
	return p.w
}

func (p *WindowsPty) OutputPipe() *os.File {
	return p.r
}

func (p *WindowsPty) Close() error {
	_ = p.r.Close()
	_ = p.w.Close()

	_ = p.consoleR.Close()
	_ = p.consoleW.Close()

	closePseudoConsole, err := kernel32DLL.FindProc("ClosePseudoConsole")
	if err != nil {
		return os.NewSyscallError("ClosePseudoConsole", err)
	}

	_, _, err = closePseudoConsole.Call(p.handle)
	return err
}

var _ Tty = (*WindowsTty)(nil)

type WindowsTty struct {
	handle uintptr
	r, w   *os.File
}

func (t *WindowsTty) Fd() uintptr {
	return t.handle
}

func (t *WindowsTty) Read(p []byte) (int, error) {
	return t.r.Read(p)
}

func (t *WindowsTty) Write(p []byte) (int, error) {
	return t.w.Write(p)
}

func (t *WindowsTty) Close() error {
	_ = t.r.Close()
	return t.w.Close()
}
