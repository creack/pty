package pty

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

const (
	I_STR   = uintptr((int32('S')<<8 | 010))
	ISPTM   = (int32('P') << 8) | 1
	UNLKPT  = (int32('P') << 8) | 1
	OWNERPT = (int32('P') << 8) | 5
)

type strioctl struct {
	ic_cmd    int32
	ic_timout int32
	ic_len    int32
	ic_dp     unsafe.Pointer
}

func ioctl(fd, cmd, ptr uintptr) error {
	return unix.IoctlSetInt(int(fd), uint(cmd), int(ptr))
}
