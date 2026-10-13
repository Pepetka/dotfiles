//go:build !linux

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "linux-backend observes the Linux desktop; use the macOS backend on this platform")
	os.Exit(1)
}
