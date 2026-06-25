package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// deleteRegistryFlags holds every flag the `delete registry` subcommand binds.
type deleteRegistryFlags struct {
	dryRunFlags
}

// DeleteRegistry builds the `registry` subcommand of `delete`.
func DeleteRegistry() *cobra.Command {
	var flags deleteRegistryFlags
	cmd := &cobra.Command{
		Use:           "registry <name>",
		Short:         "Unregister a source and cascade-uninstall its artifacts",
		Long:          "Registry removes the named registry and uninstalls every artifact it delivered. A missing registry exits successfully.",
		Args:          usageArgs(cobra.ExactArgs(1)),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteRegistry(cmd.Context(), &flags, args, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	bindDryRunFlags(cmd, &flags.dryRunFlags)

	return cmd
}

// deleteRegistry holds the cobra-free logic: it builds the input, invokes the use
// case through the fx graph, and renders the outcome to stdout.
func deleteRegistry(ctx context.Context, flags *deleteRegistryFlags, args []string, stdout io.Writer) error {
	in := usecase.DeleteRegistryInput{Name: args[0], DryRun: flags.DryRun}

	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.DeleteRegistryUseCase) error {
		result, err := uc.Execute(runCtx, in)
		if err != nil {
			return err
		}
		return renderDelete(stdout, result)
	})
}

// renderDelete writes the cascade plan groups and the summary line; a registry
// that did not exist renders only the idempotent-delete notice.
func renderDelete(stdout io.Writer, result *usecase.DeleteRegistryResult) error {
	if !result.Existed {
		_, err := fmt.Fprintf(stdout, "registry %q does not exist; nothing was deleted\n", result.Name)
		return err
	}

	groups := []Group{
		{Heading: "skills", Items: result.Plan.Skills},
		{Heading: "agents", Items: result.Plan.Agents},
		{Heading: "personas", Items: result.Plan.Personas},
	}
	if err := RenderGroups(stdout, groups); err != nil {
		return err
	}

	_, err := fmt.Fprint(stdout, deleteSummary(result))
	return err
}

// deleteSummary is the closing line: a dry-run previews the removal, an applied
// delete reports it, each with the artifact count.
func deleteSummary(result *usecase.DeleteRegistryResult) string {
	total := result.Plan.Total()
	if result.DryRun {
		return fmt.Sprintf("registry %q would be removed; %d artifacts would be removed\n", result.Name, total)
	}

	return fmt.Sprintf("registry %q removed; %d artifacts removed\n", result.Name, total)
}
