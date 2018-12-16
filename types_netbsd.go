// +build ignore

package pty

/*
#include <sys/ioctl.h>
#include <stdlib.h>
*/
import "C"

type ptmget C.struct_ptmget

var ioctl_TIOCPTMGET = C.TIOCPTMGET
