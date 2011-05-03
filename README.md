# pty

Pty is a Go package for using unix pseudo-terminals.

## Install

    goinstall github.com/kr/pty

## Example

    package main

    import (
        "fmt"
        "github.com/kr/pty"
        "io"
        "os"
    )


    func main() {
        c, err := pty.Run(
            "/bin/grep",
            []string{"grep", "--color=auto", "bar"},
            nil,
            "",
        )
        if err != nil {
            panic(err)
        }

        go func() {
            fmt.Fprintln(c.Stdin, "foo")
            fmt.Fprintln(c.Stdin, "bar")
            fmt.Fprintln(c.Stdin, "baz")
            c.Stdin.Close()
        }()
        io.Copy(os.Stdout, c.Stdout)
        c.Wait(0)
    }
