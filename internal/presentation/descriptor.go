package presentation

import (
	"fmt"
	"io"
	"strings"
)

// indentUnit is the two-space indent each nested descriptor level adds.
const indentUnit = "  "

// gap is the spacing after the longest bare "label:" in the descriptor; nested
// labels carry an indent, so their values land closer to their own colon while
// still aligning to the shared value column.
const gap = 2

// Descriptor is a kubectl-describe-style vertical view of one record: an ordered
// list of fields, each a left-aligned label with either a leaf value or a nested
// block of sub-fields. It is the single-record counterpart to Table, and a pure
// value type — no fx wiring and no knowledge of what it describes.
type Descriptor struct {
	Fields []Field
}

// Field is one descriptor entry: a label with either a leaf Value (rendered
// "label: value") or a nested block of Children (rendered "label:" followed by
// the children indented one level, e.g. an auth section). When Children is
// non-empty the Value is ignored.
type Field struct {
	Label    string
	Value    string
	Children []Field
}

// Render writes the descriptor to w. Every leaf value aligns into a single column
// — the longest bare "label:" in the tree plus a fixed gap — so a nested block's
// labels (which carry an indent) still line their values up with the top-level
// fields (kubectl-describe style). A descriptor with no fields produces no output.
func (d Descriptor) Render(w io.Writer) error {
	column := valueColumn(d.Fields)
	return renderFields(w, d.Fields, 0, column)
}

// renderFields writes one block of fields at the given indent depth, padding each
// leaf "label:" so its value begins at the shared column, and recursing into a
// section field's children one indent deeper.
func renderFields(w io.Writer, fields []Field, depth, column int) error {
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
func valueColumn(fields []Field) int {
	widest := widestLeafLabel(fields)
	if widest == 0 {
		return 0
	}

	return widest + 1 + gap
}

// widestLeafLabel is the longest leaf-field label anywhere in the tree, measured
// without indent; section labels do not participate.
func widestLeafLabel(fields []Field) int {
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
