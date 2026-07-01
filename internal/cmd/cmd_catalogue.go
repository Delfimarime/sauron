package cmd

import (
	"context"
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
// subcommands. It is a pure command group with no run behaviour: a bare
// invocation prints help and exits 0; a kind noun selects the leaf that lists.
func Catalogue() *cobra.Command {
	return newGroup(
		"catalogue",
		"Browse what the registry offers",
		"Catalogue lists the skills or agents the registry offers, fetched live from its source.",
		ListCatalogueSkill(), ListCatalogueAgent(),
	)
}

// ListCatalogueSkill builds the `skill` subcommand of `catalogue`.
func ListCatalogueSkill() *cobra.Command {
	return newCatalogueCommand(
		usecase.CatalogueSkill, "skill",
		"List the skills the registry offers",
		"Skill lists the skills the registry offers as a NAME KIND table followed by a paging line.",
	)
}

// ListCatalogueAgent builds the `agent` subcommand of `catalogue`.
func ListCatalogueAgent() *cobra.Command {
	return newCatalogueCommand(
		usecase.CatalogueAgent, "agent",
		"List the agents the registry offers",
		"Agent lists the agents the registry offers as a NAME KIND table followed by a paging line.",
	)
}

// newCatalogueCommand builds one catalogue leaf command bound to its kind; the
// leaves differ only by the kind they delegate to listCatalogue.
func newCatalogueCommand(kind usecase.CatalogueKind, use, short, long string) *cobra.Command {
	var flags catalogueFlags
	cmd := &cobra.Command{
		Use:           use,
		Short:         short,
		Long:          long,
		Args:          usageArgs(cobra.NoArgs),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return listCatalogue(cmd.Context(), kind, &flags, cmd.OutOrStdout())
		},
	}
	silenceFlagErrors(cmd)
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
// input, lets the fx graph invoke the use case, and renders the listing,
// returning the classified failure to the caller.
func listCatalogue(ctx context.Context, kind usecase.CatalogueKind, flags *catalogueFlags, stdout io.Writer) error {
	in, err := newListCatalogueInput(kind, flags)
	if err != nil {
		return err
	}

	result, err := runUseCase(ctx, func(runCtx context.Context, uc *usecase.ListCatalogueUseCase) (*usecase.ListCatalogueResponse, error) {
		return uc.Execute(runCtx, in)
	})
	if err != nil {
		return err
	}

	return renderCatalogue(stdout, result)
}

// newListCatalogueInput maps the kind and the parsed flags onto the use case's
// input, defaulting and validating the view options (--sort, --order) at this
// boundary; an invalid value yields a usage error before the use case runs.
// --search is a free substring and is not validated.
func newListCatalogueInput(kind usecase.CatalogueKind, flags *catalogueFlags) (usecase.ListCatalogueRequest, error) {
	sort := defaultCatalogueSort(flags.Sort)
	if err := validateCatalogueSort(sort); err != nil {
		return usecase.ListCatalogueRequest{}, err
	}

	order := defaultOrder(flags.Order)
	if err := validateOrder(order); err != nil {
		return usecase.ListCatalogueRequest{}, err
	}

	return usecase.ListCatalogueRequest{
		Kind:   kind,
		Search: flags.Search,
		Sort:   sort,
		Order:  order,
		Page:   flags.paging.Page,
		Limit:  flags.paging.Limit,
	}, nil
}
