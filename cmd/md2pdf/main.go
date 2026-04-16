package main

import (
	"fmt"
	"os"

	"github.com/julien/md2pdf/internal/cli"
)

func main() {
	exitCode, err := cli.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
