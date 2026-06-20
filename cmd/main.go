// Package main is the entrypoint for the sauron binary.
package main

import (
	"fmt"
	"os"

	"github.com/delfimarime/sauron/internal/cmd"
)

// Build-time identity, injected via -ldflags by `task build`.
var (
	AppName    string
	AppVersion string
	AppHash    string
)

func main() {
	root, err := cmd.New(AppName, AppVersion, AppHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(cmd.ExitCode(err))
	}
}
