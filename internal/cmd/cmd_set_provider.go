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

	if err := renderSetProvider(stdout, result); err != nil {
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
