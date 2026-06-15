package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

// listingFlags groups the filter, sort, and column flags shared by list and
// describe commands.
type listingFlags struct {
	Search string
	Sort   string
	Order  string
	Fields []string
}

func bindListingFlags(cmd *cobra.Command, f *listingFlags) {
	flags := cmd.Flags()
	flags.StringVar(&f.Search, "search", "", "case-insensitive substring filter")
	flags.StringVar(&f.Sort, "sort", "", "sort field")
	flags.StringVar(&f.Order, "order", "asc", "sort direction (asc|desc)")
	flags.StringSliceVar(&f.Fields, "fields", nil, "columns to display, in order")
}

// pagingFlags groups the offset/limit paging shared by catalogue browsing.
type pagingFlags struct {
	Offset int
	Limit  int
}

func bindPagingFlags(cmd *cobra.Command, f *pagingFlags) {
	flags := cmd.Flags()
	flags.IntVar(&f.Offset, "offset", 0, "number of leading results to skip")
	flags.IntVar(&f.Limit, "limit", 0, "maximum number of results to return (0 = all)")
}

// dryRunFlags groups the dry-run toggle shared by mutating commands.
type dryRunFlags struct {
	DryRun bool
}

func bindDryRunFlags(cmd *cobra.Command, f *dryRunFlags) {
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print the plan without changing the environment or the track file")
}

// timeoutFlags groups the network operation timeout shared by network-bound commands.
type timeoutFlags struct {
	Timeout time.Duration
}

func bindTimeoutFlags(cmd *cobra.Command, f *timeoutFlags) {
	cmd.Flags().DurationVar(&f.Timeout, "timeout", 30*time.Second, "bound on network operations")
}
