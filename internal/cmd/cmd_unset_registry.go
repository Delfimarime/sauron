package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// UnsetRegistry builds the `registry` subcommand of `unset`.
func UnsetRegistry() *cobra.Command {
	var flags dryRunFlags
	return newCommand("registry", "Remove the configured source",
		withLong("Registry removes the configured registry. Installed artifacts are preserved. With no registry configured it exits successfully."),
		withArgs(cobra.NoArgs),
		withFlags(func(cmd *cobra.Command) { bindDryRunFlags(cmd, &flags) }),
		withRunE(func(ctx context.Context, _ []string, stdout io.Writer) error {
			return unsetRegistry(ctx, &flags, stdout)
		}),
	)
}

// unsetRegistry holds the cobra-free logic: it builds the input, lets the fx
// graph invoke the use case, and renders the outcome, returning the classified
// failure to the caller.
func unsetRegistry(ctx context.Context, flags *dryRunFlags, stdout io.Writer) error {
	result, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.UnsetRegistryRequest, usecase.UnsetRegistryResponse]) (*usecase.UnsetRegistryResponse, error) {
		return uc.Execute(runCtx, usecase.UnsetRegistryRequest{DryRun: flags.DryRun})
	})
	if err != nil {
		return err
	}

	ew := newErrWriter(stdout)
	ew.printf("%s\n", unsetMessages[result.Outcome])
	return ew.toIOError("write report")
}

// unsetMessages maps each removal outcome to its canonical report line.
var unsetMessages = map[usecase.UnsetOutcome]string{
	usecase.UnsetNothing: "no registry configured; nothing was unset",
	usecase.UnsetPreview: "registry would be unset; installed artifacts preserved",
	usecase.UnsetRemoved: "registry unset; installed artifacts preserved",
}
