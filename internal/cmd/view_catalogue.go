package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// the catalogue table column headers.
const (
	headerName = "name"
	headerKind = "kind"
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

// renderCatalogue writes the name/kind table for the listed artifacts, then
// always writes the paging line. kind is the invoked command's own kind — the
// response no longer carries it back, since the caller already has it.
func renderCatalogue(w io.Writer, kind usecase.CatalogueKind, result *usecase.ListCatalogueResponse) error {
	ew := newErrWriter(w)
	ew.record(catalogueTable(kind, result).render(w))
	ew.printf("%s\n", pagingLine(result.Page, result.Limit, result.Offset, len(result.Items)))
	return ew.toIOError("render catalogue")
}

// catalogueTable builds the name/kind table; every row carries the listing kind.
func catalogueTable(kind usecase.CatalogueKind, result *usecase.ListCatalogueResponse) table {
	rows := make([][]string, len(result.Items))
	for i, name := range result.Items {
		rows[i] = []string{name, string(kind)}
	}

	return table{Headers: []string{headerName, headerKind}, Rows: rows}
}
