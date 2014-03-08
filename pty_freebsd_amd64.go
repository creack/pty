package pty

import "unsafe"

// from <sys/filio.h>
type fiodgnameArg struct {
	Len int64
	Buf uintptr
}

func newFiodgnameArg(buf []byte) *fiodgnameArg {
	return &fiodgnameArg{int64(len(buf)), uintptr(unsafe.Pointer(&buf[0]))}
}
