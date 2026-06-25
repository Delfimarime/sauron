package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

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

// listRegistries holds the cobra-free logic: it validates the view options,
// invokes the use case through the fx graph, and renders the result to stdout.
func listRegistries(ctx context.Context, flags *listingFlags, stdout io.Writer) error {
	opts := RegistryListOptions{
		Search: flags.Search,
		Sort:   flags.Sort,
		Order:  flags.Order,
		Fields: flags.Fields,
	}
	if err := opts.Validate(); err != nil {
		return usecase.NewUsageError(err.Error())
	}

	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.ListRegistriesUseCase) error {
		result, err := uc.Execute(runCtx, usecase.ListRegistriesInput{})
		if err != nil {
			return err
		}
		return RenderRegistryList(stdout, result.Registries, opts)
	})
}
