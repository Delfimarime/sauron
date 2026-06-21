package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/internal/infrastructure/repository"
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
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	request := newDescribeRegistryRequest(runCtx, flags, args, stdout)

	var execErr error
	app := NewApp(runCtx,
		repository.NewFxOptions(),
		usecase.NewFxOptions(),
		fx.Invoke(func(uc *usecase.DescribeRegistryUseCase) {
			execErr = uc.Execute(request)
		}),
	)
	if err := app.Err(); err != nil {
		return fmt.Errorf("build application: %w", err)
	}
	if err := app.Start(runCtx); err != nil {
		return fmt.Errorf("start application: %w", err)
	}
	cancel()
	_ = app.Stop(context.WithoutCancel(ctx))

	return execErr
}

// newDescribeRegistryRequest maps the positional name and parsed flags onto the
// use case's request, binding it to ctx and the command's output writer.
func newDescribeRegistryRequest(ctx context.Context, flags *describeRegistryFlags, args []string, stdout io.Writer) *usecase.DescribeRegistryRequest {
	request := usecase.NewDescribeRegistryRequest(ctx, stdout)
	request.Name = args[0]
	request.Fields = flags.Fields
	return request
}
