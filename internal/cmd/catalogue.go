package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// catalogueFlags groups the filter, sort, and paging flags the catalogue leaf
// commands share.
type catalogueFlags struct {
	Search string
	Sort   string
	Order  string
	paging pagingFlags
}

// Catalogue builds the `catalogue` command group and attaches its per-kind
// subcommands.
func Catalogue() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "catalogue",
		Short:         "Browse what a registry offers",
		Long:          "Catalogue lists the skills, agents, or personas a registry offers, fetched live from its source.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(
		ListCatalogueSkill(), ListCatalogueAgent(), ListCataloguePersona(),
	)
	return cmd
}

// ListCatalogueSkill builds the `skill` subcommand of `catalogue`.
func ListCatalogueSkill() *cobra.Command {
	return newCatalogueCommand(
		usecase.CatalogueSkill, "skill",
		"List the skills a registry offers",
		"Skill lists the skills a registry offers as a NAME KIND table followed by a paging line.",
	)
}

// ListCatalogueAgent builds the `agent` subcommand of `catalogue`.
func ListCatalogueAgent() *cobra.Command {
	return newCatalogueCommand(
		usecase.CatalogueAgent, "agent",
		"List the agents a registry offers",
		"Agent lists the agents a registry offers as a NAME KIND table followed by a paging line.",
	)
}

// ListCataloguePersona builds the `persona` subcommand of `catalogue`.
func ListCataloguePersona() *cobra.Command {
	return newCatalogueCommand(
		usecase.CataloguePersona, "persona",
		"List the personas a registry offers",
		"Persona lists the personas a registry offers as a NAME MEMBERS table followed by a paging line.",
	)
}

// newCatalogueCommand builds one catalogue leaf command bound to its kind; the
// three leaves differ only by the kind they delegate to listCatalogue.
func newCatalogueCommand(kind usecase.CatalogueKind, use, short, long string) *cobra.Command {
	var flags catalogueFlags
	cmd := &cobra.Command{
		Use:           use + " <registry>",
		Short:         short,
		Long:          long,
		Args:          usageArgs(cobra.ExactArgs(1)),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCatalogue(cmd.Context(), kind, &flags, args, cmd.OutOrStdout())
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})

	bindCatalogueFlags(cmd, &flags)

	return cmd
}

// bindCatalogueFlags registers the filter, sort, and shared paging flags.
func bindCatalogueFlags(cmd *cobra.Command, f *catalogueFlags) {
	flags := cmd.Flags()
	flags.StringVar(&f.Search, "search", "", "case-insensitive substring filter on the entry name")
	flags.StringVar(&f.Sort, "sort", "", "sort field")
	flags.StringVar(&f.Order, "order", "asc", "sort direction (asc|desc)")
	bindPagingFlags(cmd, &f.paging)
}

// listCatalogue holds the cobra-free logic shared by every kind: it builds the
// request and lets the fx graph invoke the use case, returning the classified
// failure to the caller.
func listCatalogue(ctx context.Context, kind usecase.CatalogueKind, flags *catalogueFlags, args []string, stdout io.Writer) error {
	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.ListCatalogueUseCase) error {
		return uc.Execute(newListCatalogueRequest(runCtx, kind, flags, args, stdout))
	})
}

// newListCatalogueRequest maps the kind, the positional registry, and the parsed
// flags onto the use case's request, binding it to ctx and the command's output
// writer.
func newListCatalogueRequest(ctx context.Context, kind usecase.CatalogueKind, flags *catalogueFlags, args []string, stdout io.Writer) *usecase.ListCatalogueRequest {
	request := usecase.NewListCatalogueRequest(ctx, stdout)
	request.Kind = kind
	request.Registry = args[0]
	request.Search = flags.Search
	request.Sort = flags.Sort
	request.Order = flags.Order
	request.Page = flags.paging.Page
	request.Limit = flags.paging.Limit
	return request
}
