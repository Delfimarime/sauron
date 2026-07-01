package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

const (
	// the catalogue table column headers.
	tableHeaderName = "name"
	tableHeaderKind = "kind"

	// catalogueSortName is the only field a catalogue listing sorts by; this view
	// owns the set --sort may select from.
	catalogueSortName = "name"
)

// Catalogue builds the `catalogue` command group and attaches its per-kind
// subcommands. It is a pure command group with no run behaviour: a bare
// invocation prints help and exits 0; a kind noun selects the leaf that lists.
func Catalogue() *cobra.Command {
	return newCommand("catalogue", "Browse what the registry offers",
		withLong("Catalogue lists the skills or agents the registry offers, fetched live from its source."),
		withSubcommands(
			ListCatalogueSkill(), ListCatalogueAgent(),
		),
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
// leaves differ only by the kind their run closure captures and passes to
// listCatalogue.
func newCatalogueCommand(kind usecase.CatalogueKind, use, short, long string) *cobra.Command {
	var flags listFlags
	return newCommand(use, short,
		withLong(long),
		withArgs(cobra.NoArgs),
		withFlags(func(cmd *cobra.Command) { bindListFlags(cmd, &flags, "entry") }),
		withRunE(func(ctx context.Context, _ []string, stdout io.Writer) error {
			in, err := newListCatalogueInput(kind, &flags)
			if err != nil {
				return err
			}

			result, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.ListCatalogueRequest, usecase.ListCatalogueResponse]) (*usecase.ListCatalogueResponse, error) {
				return uc.Execute(runCtx, in)
			})
			if err != nil {
				return err
			}

			ew := newErrWriter(stdout)
			rows := buildTable([]string{tableHeaderName, tableHeaderKind}, result.Items, func(name string) []string {
				return []string{name, string(kind)}
			})
			ew.record(rows.render(stdout))
			ew.printf("%s\n", pagingLine(result.Page, result.Limit, result.Offset, len(result.Items)))
			return ew.toIOError("render catalogue")
		}),
	)
}

// newListCatalogueInput maps the kind and the parsed flags onto the use case's
// input, defaulting and validating the view options (--sort, --order) at this
// boundary; an invalid value yields a usage error before the use case runs.
// --search is a free substring and is not validated.
func newListCatalogueInput(kind usecase.CatalogueKind, flags *listFlags) (usecase.ListCatalogueRequest, error) {
	sort := catalogueSortName
	if flags.Sort != "" {
		sort = flags.Sort
	}
	if sort != catalogueSortName {
		return usecase.ListCatalogueRequest{}, fmt.Errorf("%w: unknown sort field %q", errInvalidFlag, sort)
	}

	order := defaultOrder(flags.Order)
	if err := validateOrder(order); err != nil {
		return usecase.ListCatalogueRequest{}, err
	}

	return usecase.ListCatalogueRequest{
		Kind: kind,
		ListWindow: usecase.ListWindow{
			Search: flags.Search,
			Sort:   sort,
			Order:  order,
			Page:   flags.paging.Page,
			Limit:  flags.paging.Limit,
		},
	}, nil
}
