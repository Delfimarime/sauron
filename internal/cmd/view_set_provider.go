package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// renderSetProvider writes the outcome of setting the provider: the migration
// plan grouped under skills:/agents: with a summary line on a change, or a
// no-change notice when the provider was already active.
func renderSetProvider(w io.Writer, result *usecase.SetProviderResult) error {
	if result.Unchanged {
		return writeLine(w, fmt.Sprintf("provider already set to %q\n", result.Provider))
	}

	if err := renderGroup(w, "skills", result.Skills); err != nil {
		return err
	}
	if err := renderGroup(w, "agents", result.Agents); err != nil {
		return err
	}

	return writeLine(w, summaryLine(result))
}

// renderGroup writes one named plan group with a `~` marker per entry, or
// nothing when the group is empty.
func renderGroup(w io.Writer, label string, names []string) error {
	if len(names) == 0 {
		return nil
	}
	if err := writeLine(w, label+":\n"); err != nil {
		return err
	}
	for _, name := range names {
		if err := writeLine(w, fmt.Sprintf("  ~ %s\n", name)); err != nil {
			return err
		}
	}
	return nil
}

// summaryLine builds the closing confirmation, appending the migrated count only
// when at least one artifact moved.
func summaryLine(result *usecase.SetProviderResult) string {
	if result.Migrated == 0 {
		return fmt.Sprintf("provider set to %q\n", result.Provider)
	}
	return fmt.Sprintf("provider set to %q; %d artifacts migrated\n", result.Provider, result.Migrated)
}

// writeLine writes s, classifying a writer failure as an io error.
func writeLine(w io.Writer, s string) error {
	if _, err := fmt.Fprint(w, s); err != nil {
		return usecase.NewIOError(fmt.Sprintf("write report: %v", err))
	}
	return nil
}
