# pty

Pty is a Go package for using unix pseudo-terminals.

## Install

    go get github.com/kr/pty

## Examples

```go
package main

import (
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
		f.Write([]byte("foo\n"))
		f.Write([]byte("bar\n"))
		f.Write([]byte("baz\n"))
		f.Write([]byte{4}) // EOT
	}()
	io.Copy(os.Stdout, f)
}
```

For interactive mode (such as bash) use the `StartRaw` function:

```go
package main

import (
	"github.com/kr/pty"
	"io"
	"os"
	"os/exec"
)

func main() {
	c := exec.Command("bash")
	f, restore, err := pty.StartRaw(c)
	if err != nil {
		panic(err)
	}

	defer restore()

	go func() {
		io.Copy(f, os.Stdin)
	}()
	io.Copy(os.Stdout, f)
}
```
