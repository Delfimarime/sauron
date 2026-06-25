package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// describeRegistryFlags holds the flags the `describe registry` subcommand binds.
type describeRegistryFlags struct {
	Fields []string
}

// DescribeRegistry builds the `registry` subcommand of `describe`.
func DescribeRegistry() *cobra.Command {
	var flags describeRegistryFlags
	cmd := &cobra.Command{
		Use:           "registry <name>",
		Short:         "Show one registry's full detail",
		Long:          "Registry prints one registry's full detail as a vertical key-value view, with field selection.",
		Args:          usageArgs(cobra.ExactArgs(1)),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return describeRegistry(cmd.Context(), &flags, args, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	cmd.Flags().StringSliceVar(&flags.Fields, "fields", nil, "fields to display, in order; name is always first")

	return cmd
}

// describeRegistry holds the cobra-free logic: it validates the view options,
// invokes the use case through the fx graph, and renders the result to stdout.
func describeRegistry(ctx context.Context, flags *describeRegistryFlags, args []string, stdout io.Writer) error {
	opts := RegistryDetailOptions{Fields: flags.Fields}
	if err := opts.Validate(); err != nil {
		return usecase.NewUsageError(err.Error())
	}

	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.DescribeRegistryUseCase) error {
		result, err := uc.Execute(runCtx, usecase.DescribeRegistryInput{Name: args[0]})
		if err != nil {
			return err
		}
		return RenderRegistryDetail(stdout, *result, opts)
	})
}
