package cmd

import (
	"errors"
	"fmt"
)

// the sentinel causes a view option carries when it names something unknown.
var (
	errUnknownField = errors.New("unknown field")
	errUnknownSort  = errors.New("unknown sort field")
	errUnknownOrder = errors.New("unknown order")
)

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
			return nil, fmt.Errorf("%w %q", errUnknownField, f)
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
