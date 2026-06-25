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

// deleteRegistry holds the cobra-free logic: it builds the request and lets the fx
// graph invoke the use case, returning the classified failure to the caller.
func deleteRegistry(ctx context.Context, flags *deleteRegistryFlags, args []string, stdout io.Writer) error {
	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.DeleteRegistryUseCase) error {
		return uc.Execute(newDeleteRegistryRequest(runCtx, flags, args, stdout))
	})
}

// newDeleteRegistryRequest maps the parsed flags and positional argument onto the
// use case's request, binding it to ctx and the command's output writer.
func newDeleteRegistryRequest(ctx context.Context, flags *deleteRegistryFlags, args []string, stdout io.Writer) *usecase.DeleteRegistryRequest {
	request := usecase.NewDeleteRegistryRequest(ctx, stdout)
	request.Name = args[0]
	request.DryRun = flags.DryRun
	return request
}
