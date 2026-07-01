package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// describeProviderFlags holds the flags the `describe provider` subcommand binds.
type describeProviderFlags struct {
	Fields []string
}

// DescribeProvider builds the `provider` subcommand of `describe`.
func DescribeProvider() *cobra.Command {
	var flags describeProviderFlags
	cmd := &cobra.Command{
		Use:           "provider",
		Short:         "Show the active provider's full detail",
		Long:          "Provider prints the active provider's full detail as a vertical key-value view, with field selection.",
		Args:          usageArgs(cobra.NoArgs),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return describeProvider(cmd.Context(), &flags, cmd.OutOrStdout())
		},
	}
	silenceFlagErrors(cmd)
	cmd.Flags().StringSliceVar(&flags.Fields, "fields", nil, "fields to display, in order; name is always first")

	return cmd
}

// describeProvider holds the cobra-free logic: it validates the requested fields
// at this boundary, lets the fx graph invoke the use case, and renders the
// returned detail. When no provider is set the use case returns nil; the handler
// prints the none-set line and exits successfully. An unknown field yields a usage
// error before the use case runs.
func describeProvider(ctx context.Context, flags *describeProviderFlags, stdout io.Writer) error {
	fields, err := selectDescribeProviderFields(flags.Fields)
	if err != nil {
		return err
	}

	provider, err := runUseCase(ctx, func(runCtx context.Context, uc *usecase.DescribeProviderUseCase) (*types.Provider, error) {
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
