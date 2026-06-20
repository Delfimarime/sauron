package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// the transports a source may be reached over.
const (
	kindGit        = "git"
	kindHTTP       = "http"
	kindFilesystem = "filesystem"
)

// kindValues are the transports a source may be reached over.
var kindValues = []string{kindGit, kindHTTP, kindFilesystem}

// kindFlags groups the transport selector shared by source-defining commands.
type kindFlags struct {
	Kind string
}

func bindKindFlags(cmd *cobra.Command, f *kindFlags) {
	cmd.Flags().StringVar(&f.Kind, "kind", kindHTTP,
		fmt.Sprintf("source transport (%s)", strings.Join(kindValues, "|")))
}

// validateKind reports a usage error when kind is not a known transport.
func (f *kindFlags) validate() error {
	for _, v := range kindValues {
		if f.Kind == v {
			return nil
		}
	}

	return fmt.Errorf("%w: kind must be one of %s", errInvalidFlag, strings.Join(kindValues, "|"))
}

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
