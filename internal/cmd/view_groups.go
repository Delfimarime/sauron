package cmd

import (
	"fmt"
	"io"
)

// Group is a labelled list of items rendered as a heading followed by its
// "-"-prefixed entries.
type Group struct {
	Heading string
	Items   []string
}

// RenderGroups writes each non-empty group as "heading:" followed by one
// "  - item" line per entry; an empty group renders nothing.
func RenderGroups(w io.Writer, groups []Group) error {
	for _, group := range groups {
		if len(group.Items) == 0 {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s:\n", group.Heading); err != nil {
			return err
		}
		for _, item := range group.Items {
			if _, err := fmt.Fprintf(w, "  - %s\n", item); err != nil {
				return err
			}
		}
	}

	return nil
}
