package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// describeRegistryFlags holds the flags the `describe registry` subcommand binds.
type describeRegistryFlags struct {
	Fields []string
}

// DescribeRegistry builds the `registry` subcommand of `describe`.
func DescribeRegistry() *cobra.Command {
	var flags describeRegistryFlags
	cmd := &cobra.Command{
		Use:           "registry",
		Short:         "Show the configured registry's full detail",
		Long:          "Registry prints the configured registry's full detail as a vertical key-value view, with field selection.",
		Args:          usageArgs(cobra.NoArgs),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return describeRegistry(cmd.Context(), &flags, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	cmd.Flags().StringSliceVar(&flags.Fields, "fields", nil, "fields to display, in order; uri is always first")

	return cmd
}

// describeRegistry holds the cobra-free logic: it validates the requested fields
// at this boundary, lets the fx graph invoke the use case, and renders the
// returned detail, returning the classified failure to the caller. An unknown
// field yields a usage error before the use case runs.
func describeRegistry(ctx context.Context, flags *describeRegistryFlags, stdout io.Writer) error {
	fields, err := selectDescribeFields(flags.Fields)
	if err != nil {
		return err
	}

	registry, err := runUseCase(ctx, func(runCtx context.Context, uc *usecase.DescribeRegistryUseCase) (*types.Registry, error) {
		return uc.Execute(runCtx, usecase.DescribeRegistryInput{})
	})
	if err != nil {
		return err
	}

	return renderDescribeRegistry(stdout, registry, fields)
}
