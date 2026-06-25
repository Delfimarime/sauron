package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// unsetRegistryFlags holds every flag the `unset registry` subcommand binds.
type unsetRegistryFlags struct {
	dryRunFlags
}

// UnsetRegistry builds the `registry` subcommand of `unset`.
func UnsetRegistry() *cobra.Command {
	var flags unsetRegistryFlags
	cmd := &cobra.Command{
		Use:           "registry",
		Short:         "Remove the configured source",
		Long:          "Registry removes the configured registry. Installed artifacts are preserved. With no registry configured it exits successfully.",
		Args:          usageArgs(cobra.NoArgs),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return unsetRegistry(cmd.Context(), &flags, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	bindDryRunFlags(cmd, &flags.dryRunFlags)

	return cmd
}

// unsetRegistry holds the cobra-free logic: it builds the input, lets the fx
// graph invoke the use case, and renders the outcome, returning the classified
// failure to the caller.
func unsetRegistry(ctx context.Context, flags *unsetRegistryFlags, stdout io.Writer) error {
	result, err := runUseCase(ctx, func(runCtx context.Context, uc *usecase.UnsetRegistryUseCase) (*usecase.UnsetRegistryResult, error) {
		return uc.Execute(runCtx, usecase.UnsetRegistryInput{DryRun: flags.DryRun})
	})
	if err != nil {
		return err
	}

	return renderUnsetRegistry(stdout, result)
}
