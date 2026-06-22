package usecase

import (
	"context"
	"fmt"
	"io"
)

// baseRequest carries the per-invocation context and output writer shared by
// every request type, satisfying the Out half of the Request interface.
type baseRequest struct {
	context.Context
	out io.Writer
}

// Out returns the writer the command's output goes to.
func (r baseRequest) Out() io.Writer { return r.out }

// selectFields validates requested against known, forcing fieldName present and
// first and deduping; an empty request yields dflt.
func selectFields(requested []string, known map[string]struct{}, dflt []string) ([]string, error) {
	if len(requested) == 0 {
		return dflt, nil
	}

	fields := []string{fieldName}
	seen := map[string]struct{}{fieldName: {}}
	for _, f := range requested {
		if _, ok := known[f]; !ok {
			return nil, NewUsageError(fmt.Sprintf("unknown field %q", f))
		}
		if _, dup := seen[f]; dup {
			continue
		}
		seen[f] = struct{}{}
		fields = append(fields, f)
	}

	return fields, nil
}

// projectRows builds table rows by applying the column projector for each
// selected column to every item.
func projectRows[T any](items []T, columns []string, projectors map[string]func(T) string) [][]string {
	out := make([][]string, len(items))
	for i, item := range items {
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = projectors[col](item)
		}
		out[i] = row
	}

	return out
}

// defaultSortOrder applies the listing defaults: an empty sort becomes fieldName
// and an empty order becomes orderAsc.
func defaultSortOrder(sort, order string) (string, string) {
	if sort == "" {
		sort = fieldName
	}
	if order == "" {
		order = orderAsc
	}

	return sort, order
}

// isValidOrder reports whether order is one of the accepted sort directions.
func isValidOrder(order string) bool {
	return order == orderAsc || order == orderDesc
}

// the selectable columns of a registry listing and describe detail.
const (
	fieldName                 = "name"
	fieldTransport            = "transport"
	fieldURI                  = "uri"
	fieldRef                  = "ref"
	fieldAuth                 = "auth"
	fieldTLS                  = "tls"
	fieldSSHKey               = "sshKey"
	fieldTimeout              = "timeout"
	fieldCreationTimestamp    = "creationTimestamp"
	fieldLastUpdatedTimestamp = "lastUpdatedTimestamp"
)

// the sort directions a listing accepts.
const (
	orderAsc  = "asc"
	orderDesc = "desc"
)

// predicate reports whether an item should be kept.
type predicate[T any] func(T) bool

// filterBy keeps the items the predicate accepts.
func filterBy[T any](items []T, keep predicate[T]) []T {
	kept := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			kept = append(kept, item)
		}
	}

	return kept
}
