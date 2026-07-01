// Package main is the entrypoint for the sauron binary.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

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

	os.Exit(run(root))
}

// run executes root under a graceful-shutdown context and returns the exit
// code. Using a separate function means defer stop() always fires before
// os.Exit is reached in the caller.
func run(root *cobra.Command) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return cmd.ExitCode(err)
	}

	return 0
}
