package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

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
