package presentation

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// emptyCell is the placeholder rendered for an absent optional value.
const emptyCell = "—"

// Table is an aligned, uppercase-headed table of string cells. A zero-row table
// renders nothing at all — not even its header.
type Table struct {
	Headers []string
	Rows    [][]string
}

// Render writes the table to w with aligned columns and uppercase headers,
// substituting "—" for an empty cell. A table with no rows produces no output.
func (t Table) Render(w io.Writer) error {
	if len(t.Rows) == 0 {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if err := writeRow(tw, header(t.Headers)); err != nil {
		return err
	}
	for _, row := range t.Rows {
		if err := writeRow(tw, cells(row)); err != nil {
			return err
		}
	}

	return tw.Flush()
}

// header upper-cases every column title.
func header(headers []string) []string {
	out := make([]string, len(headers))
	for i, h := range headers {
		out[i] = strings.ToUpper(h)
	}
	return out
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
