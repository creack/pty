// +build freebsd,386 freebsd,arm

package pty

import "unsafe"

// from <sys/filio.h>
type fiodgnameArg struct {
	Len int32
	Buf uintptr
}

func newFiodgnameArg(buf []byte) *fiodgnameArg {
	return &fiodgnameArg{int32(len(buf)), uintptr(unsafe.Pointer(&buf[0]))}
}
