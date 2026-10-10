package main

import (
	"dotfiles/theme/internal/install"
	"flag"
	"fmt"
	"os"
)

func main() {
	flags := flag.NewFlagSet("theme-install", flag.ExitOnError)
	source := flags.String("source", ".", "theme source directory")
	workingDirectory := flags.String("working-dir", "", "resolve relative paths from this directory")
	flags.Parse(os.Args[1:])
	if *workingDirectory != "" {
		if err := os.Chdir(*workingDirectory); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := install.Execute(*source, flags.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
