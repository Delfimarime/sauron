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

// describeRegistry holds the cobra-free logic: it builds the request and lets the
// fx graph invoke the use case, returning the classified failure to the caller.
func describeRegistry(ctx context.Context, flags *describeRegistryFlags, args []string, stdout io.Writer) error {
	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.DescribeRegistryUseCase) error {
		return uc.Execute(newDescribeRegistryRequest(runCtx, flags, args, stdout))
	})
}

// newDescribeRegistryRequest maps the positional name and parsed flags onto the
// use case's request, binding it to ctx and the command's output writer.
func newDescribeRegistryRequest(ctx context.Context, flags *describeRegistryFlags, args []string, stdout io.Writer) *usecase.DescribeRegistryRequest {
	request := usecase.NewDescribeRegistryRequest(ctx, stdout)
	request.Name = args[0]
	request.Fields = flags.Fields
	return request
}
