package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// kindHeadings maps each artifact kind to its plural output heading token.
var kindHeadings = map[string]string{
	types.KindSkill: "skills",
	types.KindAgent: "agents",
}

// renderInstall writes the install plan to w: a kind heading, one line per
// added ("+") or updated ("~") artifact, one line per failure, and a summary
// count when at least one artifact was acted on.
func renderInstall(w io.Writer, kind string, result *usecase.InstallResponse) error {
	ew := newErrWriter(w)
	ew.printf("%s:\n", kindHeadings[kind])
	for _, a := range result.Added {
		ew.printf("  + sauron-%s\n", a.Metadata.Name)
	}
	for _, a := range result.Updated {
		ew.printf("  ~ sauron-%s\n", a.Metadata.Name)
	}
	for _, f := range result.Failures {
		ew.printf("  ! %s: %s\n", f.Name, f.Reason)
	}
	if summary := installSummary(result); summary != "" {
		ew.printf("%s\n", summary)
	}

	return ew.toIOError("write install report")
}

// installSummary produces the tail summary line; an empty string is returned
// when neither added nor updated counts are non-zero (a no-op run).
func installSummary(result *usecase.InstallResponse) string {
	added := len(result.Added)
	updated := len(result.Updated)

	switch {
	case added > 0 && updated > 0:
		return fmt.Sprintf("%d added, %d updated", added, updated)
	case added > 0:
		return fmt.Sprintf("%d added", added)
	case updated > 0:
		return fmt.Sprintf("%d updated", updated)
	default:
		return ""
	}
}
