package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// the catalogue table column headers.
const (
	headerName = "NAME"
	headerKind = "KIND"
)

// catalogueSortName is the only field a catalogue listing sorts by; this view
// owns the set --sort may select from.
const catalogueSortName = "name"

// defaultCatalogueSort applies the catalogue default: an empty selection sorts by
// name.
func defaultCatalogueSort(sort string) string {
	if sort == "" {
		return catalogueSortName
	}

	return sort
}

// validateCatalogueSort reports a usage error when sort is not a sortable
// catalogue field, raised before the use case runs.
func validateCatalogueSort(sort string) error {
	if sort == catalogueSortName {
		return nil
	}

	return fmt.Errorf("%w: unknown sort field %q", errInvalidFlag, sort)
}

// renderCatalogue writes the NAME/KIND table for the listed artifacts, then
// always writes the paging line.
func renderCatalogue(w io.Writer, result *usecase.ListCatalogueResult) error {
	if err := catalogueTable(result).render(w); err != nil {
		return usecase.NewIOError(fmt.Sprintf("render table: %v", err))
	}
	if _, err := fmt.Fprintln(w, pagingLine(result)); err != nil {
		return usecase.NewIOError(fmt.Sprintf("write paging line: %v", err))
	}

	return nil
}

// catalogueTable builds the NAME/KIND table; every row carries the listing kind.
func catalogueTable(result *usecase.ListCatalogueResult) table {
	rows := make([][]string, len(result.Items))
	for i, name := range result.Items {
		rows[i] = []string{name, string(result.Kind)}
	}

	return table{Headers: []string{headerName, headerKind}, Rows: rows}
}

// pagingLine renders the applied-paging report; an empty page reports zero
// results, a populated page the inclusive from–to window.
func pagingLine(result *usecase.ListCatalogueResult) string {
	count := len(result.Items)
	if count == 0 {
		return fmt.Sprintf("showing 0 results (page %d, limit %d)", result.Page, result.Limit)
	}

	from := result.Offset + 1
	to := result.Offset + int64(count)

	return fmt.Sprintf("showing %d–%d (page %d, limit %d)", from, to, result.Page, result.Limit)
}
