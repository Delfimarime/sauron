package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// SetProvider builds the `provider` subcommand of `set`.
func SetProvider() *cobra.Command {
	return &cobra.Command{
		Use:           "provider <claude|zencoder>",
		Short:         "Set the provider artifacts are installed for",
		Long:          "Provider records the single global destination; changing it migrates every installed artifact to the new provider's directories.",
		Args:          usageArgs(cobra.ExactArgs(1)),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return setProvider(cmd.Context(), args, cmd.OutOrStdout())
		},
	}
}

// setProvider holds the cobra-free logic: it builds the input, lets the fx graph
// invoke the use case, and renders the returned result, returning any classified
// failure to the caller.
func setProvider(ctx context.Context, args []string, stdout io.Writer) error {
	result, err := runUseCase(ctx, func(runCtx context.Context, uc *usecase.SetProviderUseCase) (*usecase.SetProviderResult, error) {
		return uc.Execute(runCtx, usecase.SetProviderInput{Provider: args[0]})
	})
	if err != nil {
		return err
	}

	return renderSetProvider(stdout, result)
}
