# pty

Pty is a Go package for using unix pseudo-terminals.

(Note, there is only a Linux implementation. I'd appreciate a patch
for other systems!)

## Install

    go get github.com/kr/pty

## Example

```go
package main

import (
	"fmt"
	"github.com/kr/pty"
	"io"
	"os"
	"os/exec"
)

func main() {
	c := exec.Command("grep", "--color=auto", "bar")
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	go func() {
		fmt.Fprintln(f, "foo")
		fmt.Fprintln(f, "bar")
		fmt.Fprintln(f, "baz")
		f.Close()
	}()
	io.Copy(os.Stdout, f)
	c.Wait()
}
```
