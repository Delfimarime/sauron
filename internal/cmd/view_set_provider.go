package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// renderSetProvider writes the outcome of setting the provider: the migration
// plan grouped under skills:/agents: with any migration failures and a summary
// line on a change, or a no-change notice when the provider was already active.
func renderSetProvider(w io.Writer, result *usecase.SetProviderResponse) error {
	ew := newErrWriter(w)

	if result.Unchanged {
		ew.printf("provider already set to %q\n", result.Provider)
		return ew.toIOError("write report")
	}

	renderGroupInto(ew, "skills", result.Skills)
	renderGroupInto(ew, "agents", result.Agents)

	for _, f := range result.Failures {
		ew.printf("  ! %s: %s\n", f.Artifact.Metadata.Name, f.Reason)
	}

	ew.printf("%s", summaryLine(result))
	return ew.toIOError("write report")
}

// renderGroupInto writes one named plan group with a `~` marker per entry into
// ew, or does nothing when the group is empty.
func renderGroupInto(ew *errWriter, label string, names []string) {
	if len(names) == 0 {
		return
	}
	ew.printf("%s:\n", label)
	for _, name := range names {
		ew.printf("  ~ %s\n", name)
	}
}

// summaryLine builds the closing confirmation, appending the migrated count only
// when at least one artifact moved.
func summaryLine(result *usecase.SetProviderResponse) string {
	if result.Migrated == 0 {
		return fmt.Sprintf("provider set to %q\n", result.Provider)
	}
	return fmt.Sprintf("provider set to %q; %d artifacts migrated\n", result.Provider, result.Migrated)
}
