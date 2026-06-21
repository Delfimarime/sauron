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

// ListRegistries builds the `registries` subcommand of `list`.
func ListRegistries() *cobra.Command {
	var flags listingFlags
	cmd := &cobra.Command{
		Use:           "registries",
		Short:         "List the registered sources",
		Long:          "Registries prints the registered sources as a table, with filtering, column selection, and sorting.",
		Args:          usageArgs(cobra.NoArgs),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return listRegistries(cmd.Context(), &flags, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	bindListingFlags(cmd, &flags)

	return cmd
}

// listRegistries holds the cobra-free logic: it builds the request and lets the
// fx graph invoke the use case, returning the classified failure to the caller.
func listRegistries(ctx context.Context, flags *listingFlags, stdout io.Writer) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	request := newListRegistriesRequest(runCtx, flags, stdout)

	var execErr error
	app := NewApp(runCtx,
		repository.NewFxOptions(),
		usecase.NewFxOptions(),
		fx.Invoke(func(uc *usecase.ListRegistriesUseCase) {
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

// newListRegistriesRequest maps the parsed flags onto the use case's request,
// binding it to ctx and the command's output writer.
func newListRegistriesRequest(ctx context.Context, flags *listingFlags, stdout io.Writer) *usecase.ListRegistriesRequest {
	request := usecase.NewListRegistriesRequest(ctx, stdout)
	request.Search = flags.Search
	request.Sort = flags.Sort
	request.Order = flags.Order
	request.Fields = flags.Fields
	return request
}
