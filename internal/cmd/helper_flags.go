package cmd

import (
	"fmt"
	"slices"
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
	if slices.Contains(kindValues, f.Kind) {
		return nil
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

// the sort directions a listing accepts.
const (
	orderAsc  = "asc"
	orderDesc = "desc"
)

// defaultOrder applies the listing default: an empty order becomes asc.
func defaultOrder(order string) string {
	if order == "" {
		return orderAsc
	}

	return order
}

// validateOrder reports a usage error when order is not an accepted direction.
func validateOrder(order string) error {
	if order == orderAsc || order == orderDesc {
		return nil
	}

	return fmt.Errorf("%w: order must be %s or %s", errInvalidFlag, orderAsc, orderDesc)
}

// the default paging window shared by paginated listings.
const (
	defaultPage  = 1
	defaultLimit = 20
)

// pagingFlags groups the 1-based page number and page size shared by paginated
// listing commands.
type pagingFlags struct {
	Page  int64
	Limit int64
}

func bindPagingFlags(cmd *cobra.Command, f *pagingFlags) {
	flags := cmd.Flags()
	flags.Int64Var(&f.Page, "page", defaultPage, "1-based page number")
	flags.Int64Var(&f.Limit, "limit", defaultLimit, "page size")
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
