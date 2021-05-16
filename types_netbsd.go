// +build ignore

package pty

/*
#include <sys/time.h>
#include <stdlib.h>
#include <sys/tty.h>
*/
import "C"

type ptmget C.struct_ptmget

var ioctl_TIOCPTSNAME = C.TIOCPTSNAME
var ioctl_TIOCGRANTPT = C.TIOCGRANTPT
