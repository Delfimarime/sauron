package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DescribeRegistry builds the `registry` subcommand of `describe`.
func DescribeRegistry() *cobra.Command {
	var flags fieldsFlags
	return newCommand("registry", "Show the configured registry's full detail",
		withLong("Registry prints the configured registry's full detail as a vertical key-value view, with field selection."),
		withArgs(cobra.NoArgs),
		withFlags(func(cmd *cobra.Command) { bindFieldsFlags(cmd, &flags, "source") }),
		withRunE(func(ctx context.Context, _ []string, stdout io.Writer) error {
			return describeRegistry(ctx, &flags, stdout)
		}),
	)
}

// describeRegistry holds the cobra-free logic: it validates the requested fields
// at this boundary, lets the fx graph invoke the use case, and renders the
// returned detail, returning the classified failure to the caller. An unknown
// field yields a usage error before the use case runs.
func describeRegistry(ctx context.Context, flags *fieldsFlags, stdout io.Writer) error {
	fields, err := selectDescribeFields(flags.Fields)
	if err != nil {
		return err
	}

	registry, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.DescribeRegistryRequest, types.Registry]) (*types.Registry, error) {
		return uc.Execute(runCtx, usecase.DescribeRegistryRequest{})
	})
	if err != nil {
		return err
	}

	return renderDescribeRegistry(stdout, registry, fields)
}
