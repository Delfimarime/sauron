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

// deleteRegistryFlags holds every flag the `delete registry` subcommand binds.
type deleteRegistryFlags struct {
	dryRunFlags
}

// DeleteRegistry builds the `registry` subcommand of `delete`.
func DeleteRegistry() *cobra.Command {
	var flags deleteRegistryFlags
	cmd := &cobra.Command{
		Use:           "registry <name>",
		Short:         "Unregister a source and cascade-uninstall its artifacts",
		Long:          "Registry removes the named registry and uninstalls every artifact it delivered. A missing registry exits successfully.",
		Args:          usageArgs(cobra.ExactArgs(1)),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteRegistry(cmd.Context(), &flags, args, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	bindDryRunFlags(cmd, &flags.dryRunFlags)

	return cmd
}

// deleteRegistry holds the cobra-free logic: it builds the request and lets the fx
// graph invoke the use case, returning the classified failure to the caller.
func deleteRegistry(ctx context.Context, flags *deleteRegistryFlags, args []string, stdout io.Writer) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	request := newDeleteRegistryRequest(runCtx, flags, args, stdout)

	var execErr error
	app := NewApp(runCtx,
		repository.NewFxOptions(),
		usecase.NewFxOptions(),
		fx.Invoke(func(uc *usecase.DeleteRegistryUseCase) {
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

// newDeleteRegistryRequest maps the parsed flags and positional argument onto the
// use case's request, binding it to ctx and the command's output writer.
func newDeleteRegistryRequest(ctx context.Context, flags *deleteRegistryFlags, args []string, stdout io.Writer) *usecase.DeleteRegistryRequest {
	request := usecase.NewDeleteRegistryRequest(ctx, stdout)
	request.Name = args[0]
	request.DryRun = flags.DryRun
	return request
}
