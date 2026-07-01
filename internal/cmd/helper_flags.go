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
	transportGit  = "git"
	transportHTTP = "http"
)

// transportValues are the transports a source may be reached over.
var transportValues = []string{transportGit, transportHTTP}

// transportFlags groups the transport selector shared by source-defining commands.
type transportFlags struct {
	Transport string
}

func bindTransportFlags(cmd *cobra.Command, f *transportFlags) {
	cmd.Flags().StringVar(&f.Transport, "transport", transportHTTP,
		fmt.Sprintf("source transport (%s)", strings.Join(transportValues, "|")))
}

// validate reports a usage error when transport is not a known transport.
func (f *transportFlags) validate() error {
	if slices.Contains(transportValues, f.Transport) {
		return nil
	}

	return fmt.Errorf("%w: transport must be one of %s", errInvalidFlag, strings.Join(transportValues, "|"))
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

// listFlags groups the search/sort/order/paging flags shared by every listing
// command (catalogue, and any future list command).
type listFlags struct {
	Search string
	Sort   string
	Order  string
	paging pagingFlags
}

// bindListFlags registers --search, --sort, --order, and the shared paging
// flags; noun names what --search filters, used only in its help text (e.g.
// "entry" for catalogue).
func bindListFlags(cmd *cobra.Command, f *listFlags, noun string) {
	flags := cmd.Flags()
	flags.StringVar(&f.Search, "search", "", fmt.Sprintf("case-insensitive substring filter on the %s name", noun))
	flags.StringVar(&f.Sort, "sort", "", "sort field")
	flags.StringVar(&f.Order, "order", "asc", "sort direction (asc|desc)")
	bindPagingFlags(cmd, &f.paging)
}

// fieldsFlags groups the column/field-selection flag shared by every
// describe/list command that supports field selection.
type fieldsFlags struct {
	Fields []string
}

// bindFieldsFlags registers --fields; identity names the field that is always
// present and first, used only in the flag's help text.
func bindFieldsFlags(cmd *cobra.Command, f *fieldsFlags, identity string) {
	cmd.Flags().StringSliceVar(&f.Fields, "fields", nil,
		fmt.Sprintf("fields to display, in order; %s is always first", identity))
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
