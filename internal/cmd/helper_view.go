package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/delfimarime/sauron/internal/usecase"
)

// errWriter is a write helper that accumulates the first error and silently
// no-ops every subsequent write, so renderers can issue multiple printf calls
// without per-call error checks. Call toIOError once at the end.
type errWriter struct {
	w   io.Writer
	err error
}

// newErrWriter wraps w in a sticky-error writer.
func newErrWriter(w io.Writer) *errWriter {
	return &errWriter{w: w}
}

// printf formats and writes like fmt.Fprintf; once an error has been recorded
// all subsequent calls are silent no-ops.
func (ew *errWriter) printf(format string, args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, args...)
}

// record notes an external error (e.g. from a sub-renderer) without writing;
// once an error is recorded all subsequent printf calls are silent no-ops.
func (ew *errWriter) record(err error) {
	if ew.err == nil {
		ew.err = err
	}
}

// toIOError returns nil when no write has failed, or a classified io error
// with context prepended.
func (ew *errWriter) toIOError(context string) error {
	if ew.err == nil {
		return nil
	}
	return usecase.NewIOError(fmt.Sprintf("%s: %v", context, ew.err))
}

// pagingLine renders the applied-paging report shared by every paginated
// listing view: an empty page reports zero results, a populated page the
// inclusive from–to window.
func pagingLine(page, limit, offset int64, count int) string {
	if count == 0 {
		return fmt.Sprintf("showing 0 results (page %d, limit %d)", page, limit)
	}

	from := offset + 1
	to := offset + int64(count)

	return fmt.Sprintf("showing %d–%d (page %d, limit %d)", from, to, page, limit)
}

// emptyCell is the placeholder rendered for an absent optional value.
const emptyCell = "—"

// table is an aligned table of string cells with verbatim headers. A zero-row
// table renders nothing at all — not even its header.
type table struct {
	Headers []string
	Rows    [][]string
}

// render writes the table to w with aligned columns and verbatim headers,
// substituting "—" for an empty cell. A table with no rows produces no output.
func (t table) render(w io.Writer) error {
	if len(t.Rows) == 0 {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if err := writeRow(tw, t.Headers); err != nil {
		return err
	}
	for _, row := range t.Rows {
		if err := writeRow(tw, cells(row)); err != nil {
			return err
		}
	}

	return tw.Flush()
}

// cells substitutes the empty-cell placeholder for any blank value.
func cells(row []string) []string {
	out := make([]string, len(row))
	for i, c := range row {
		if c == "" {
			out[i] = emptyCell
			continue
		}
		out[i] = c
	}
	return out
}

// writeRow emits one tab-separated row terminated by a newline.
func writeRow(w io.Writer, cols []string) error {
	_, err := fmt.Fprintln(w, strings.Join(cols, "\t"))
	return err
}

// buildTable maps items onto a table's rows via row, one call per item, under
// the given headers — the shared "slice of T to [][]string" shape every
// table-backed view composes instead of hand-rolling its own loop.
func buildTable[T any](headers []string, items []T, row func(T) []string) table {
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = row(item)
	}

	return table{Headers: headers, Rows: rows}
}

// indentUnit is the two-space indent each nested descriptor level adds.
const indentUnit = "  "

// descriptorGap is the spacing after the longest bare "label:" in the
// descriptor; nested labels carry an indent, so their values land closer to
// their own colon while still aligning to the shared value column.
const descriptorGap = 2

// descriptor is a kubectl-describe-style vertical view of one record: an ordered
// list of fields, each a left-aligned label with either a leaf value or a nested
// block of sub-fields. It is a pure value type — no knowledge of what it
// describes.
type descriptor struct {
	Fields []descriptorField
}

// descriptorField is one descriptor entry: a label with either a leaf Value
// (rendered "label: value") or a nested block of Children (rendered "label:"
// followed by the children indented one level, e.g. an auth section). When
// Children is non-empty the Value is ignored.
type descriptorField struct {
	Label    string
	Value    string
	Children []descriptorField
}

// render writes the descriptor to w. Every leaf value aligns into a single column
// — the longest bare "label:" in the tree plus a fixed gap — so a nested block's
// labels (which carry an indent) still line their values up with the top-level
// fields (kubectl-describe style). A descriptor with no fields produces no output.
func (d descriptor) render(w io.Writer) error {
	column := valueColumn(d.Fields)
	return renderFields(w, d.Fields, 0, column)
}

// renderFields writes one block of fields at the given indent depth, padding each
// leaf "label:" so its value begins at the shared column, and recursing into a
// section field's children one indent deeper.
func renderFields(w io.Writer, fields []descriptorField, depth, column int) error {
	indent := indentOf(depth)
	for _, field := range fields {
		if len(field.Children) > 0 {
			if _, err := fmt.Fprintf(w, "%s%s:\n", indent, field.Label); err != nil {
				return err
			}
			if err := renderFields(w, field.Children, depth+1, column); err != nil {
				return err
			}
			continue
		}

		pad := column - len(indent) - len(field.Label) - 1
		if _, err := fmt.Fprintf(w, "%s%s:%s%s\n", indent, field.Label, spaces(pad), field.Value); err != nil {
			return err
		}
	}

	return nil
}

// valueColumn is the column every leaf value starts at: the longest bare "label:"
// across all leaf fields (ignoring indent) plus the fixed gap. Zero when there is
// no leaf field.
func valueColumn(fields []descriptorField) int {
	widest := widestLeafLabel(fields)
	if widest == 0 {
		return 0
	}

	return widest + 1 + descriptorGap
}

// widestLeafLabel is the longest leaf-field label anywhere in the tree, measured
// without indent; section labels do not participate.
func widestLeafLabel(fields []descriptorField) int {
	widest := 0
	for _, field := range fields {
		if len(field.Children) > 0 {
			if nested := widestLeafLabel(field.Children); nested > widest {
				widest = nested
			}
			continue
		}
		if n := len(field.Label); n > widest {
			widest = n
		}
	}

	return widest
}

// indentOf is the leading whitespace for a nesting depth.
func indentOf(depth int) string {
	return strings.Repeat(indentUnit, depth)
}

// spaces returns n spaces, or a single space when n is non-positive so a label as
// long as the column still separates from its value.
func spaces(n int) string {
	if n < 1 {
		n = 1
	}

	return strings.Repeat(" ", n)
}

// leafField builds a leaf field, reporting false for an empty value.
func leafField(label, value string) (descriptorField, bool) {
	if value == "" {
		return descriptorField{}, false
	}

	return descriptorField{Label: label, Value: value}, true
}

// sectionField builds a section field, reporting false when it has no children.
func sectionField(label string, children []descriptorField) (descriptorField, bool) {
	if len(children) == 0 {
		return descriptorField{}, false
	}

	return descriptorField{Label: label, Children: children}, true
}

// selectFields is the shared field-selector: it validates the requested fields
// against order, forces first present and first, dedupes, and returns every
// field in order for an empty request. An unknown field is a usage error (exit 2)
// raised before the use case runs.
func selectFields(requested, order []string, first string) ([]string, error) {
	if len(requested) == 0 {
		return order, nil
	}

	known := make(map[string]struct{}, len(order))
	for _, f := range order {
		known[f] = struct{}{}
	}

	fields := []string{first}
	seen := map[string]struct{}{first: {}}
	for _, f := range requested {
		if _, ok := known[f]; !ok {
			return nil, fmt.Errorf("%w: unknown field %q", errInvalidFlag, f)
		}
		if _, dup := seen[f]; dup {
			continue
		}
		seen[f] = struct{}{}
		fields = append(fields, f)
	}

	return fields, nil
}
