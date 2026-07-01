package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DescribeProvider builds the `provider` subcommand of `describe`.
func DescribeProvider() *cobra.Command {
	var flags fieldsFlags
	return newCommand("provider", "Show the active provider's full detail",
		withLong("Provider prints the active provider's full detail as a vertical key-value view, with field selection."),
		withArgs(cobra.NoArgs),
		withFlags(func(cmd *cobra.Command) { bindFieldsFlags(cmd, &flags, "name") }),
		withRunE(func(ctx context.Context, _ []string, stdout io.Writer) error {
			return describeProvider(ctx, &flags, stdout)
		}),
	)
}

// describeProvider holds the cobra-free logic: it validates the requested fields
// at this boundary, lets the fx graph invoke the use case, and renders the
// returned detail. When no provider is set the use case returns nil; the handler
// prints the none-set line and exits successfully. An unknown field yields a usage
// error before the use case runs.
func describeProvider(ctx context.Context, flags *fieldsFlags, stdout io.Writer) error {
	fields, err := selectDescribeProviderFields(flags.Fields)
	if err != nil {
		return err
	}

	provider, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.DescribeProviderRequest, types.Provider]) (*types.Provider, error) {
		return uc.Execute(runCtx, usecase.DescribeProviderRequest{})
	})
	if err != nil {
		return err
	}
	if provider == nil {
		return renderNoProvider(stdout)
	}

	return renderDescribeProvider(stdout, provider, fields)
}
