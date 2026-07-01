package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// SetProvider builds the `provider` subcommand of `set`.
func SetProvider() *cobra.Command {
	return newCommand("provider <claude|zencoder>", "Set the provider artifacts are installed for",
		withLong("Provider records the single global destination; changing it migrates every installed artifact to the new provider's directories."),
		withArgs(cobra.ExactArgs(1)),
		withRunE(func(ctx context.Context, args []string, stdout io.Writer) error {
			return setProvider(ctx, args, stdout)
		}),
	)
}

// setProvider holds the cobra-free logic: it builds the input, lets the fx graph
// invoke the use case, renders the returned result, and then checks for
// migration failures. The provider is persisted regardless (FR-005 reports and
// continues), but a stranded artifact makes the process exit 1.
func setProvider(ctx context.Context, args []string, stdout io.Writer) error {
	result, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.SetProviderRequest, usecase.SetProviderResponse]) (*usecase.SetProviderResponse, error) {
		return uc.Execute(runCtx, usecase.SetProviderRequest{Provider: args[0]})
	})
	if err != nil {
		return err
	}

	ew := newErrWriter(stdout)
	if result.Unchanged {
		ew.printf("provider already set to %q\n", result.Provider)
	} else {
		renderGroupInto(ew, "skills", result.Skills)
		renderGroupInto(ew, "agents", result.Agents)
		for _, f := range result.Failures {
			ew.printf("  ! %s: %s\n", f.Artifact.Metadata.Name, f.Reason)
		}
		ew.printf("%s", summaryLine(result))
	}
	if err := ew.toIOError("write report"); err != nil {
		return err
	}

	// A migration failure means at least one tracked artifact could not be moved
	// to the new provider's directories; the provider is still set and every
	// failure is rendered, but exit non-zero so the caller can detect and retry.
	if len(result.Failures) > 0 {
		return usecase.NewIOError(fmt.Sprintf("migrate to %q: %d artifacts stranded", result.Provider, len(result.Failures)))
	}

	return nil
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
