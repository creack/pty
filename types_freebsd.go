// +build ignore

package pty

// #include <sys/filio.h>
import "C"

type fiodgnameArg C.struct_fiodgname_arg
