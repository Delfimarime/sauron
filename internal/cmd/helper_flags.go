package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

// listingFlags groups flags shaping list-style output.
type listingFlags struct {
	Output string
	Quiet  bool
}

func bindListingFlags(cmd *cobra.Command, f *listingFlags) {
	flags := cmd.Flags()
	flags.StringVarP(&f.Output, "output", "o", "table", "output format")
	flags.BoolVarP(&f.Quiet, "quiet", "q", false, "suppress non-essential output")
}

// dryRunFlags groups the dry-run toggle shared by mutating commands.
type dryRunFlags struct {
	DryRun bool
}

func bindDryRunFlags(cmd *cobra.Command, f *dryRunFlags) {
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "report changes without applying them")
}

// timeoutFlags groups the operation timeout shared by network-bound commands.
type timeoutFlags struct {
	Timeout time.Duration
}

func bindTimeoutFlags(cmd *cobra.Command, f *timeoutFlags) {
	cmd.Flags().DurationVar(&f.Timeout, "timeout", 0, "operation timeout (0 = none)")
}
