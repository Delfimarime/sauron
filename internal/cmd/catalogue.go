package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
)

// the catalogue table column headers.
const (
	colName    = "NAME"
	colKind    = "KIND"
	colMembers = "MEMBERS"
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
// input, invokes the use case through the fx graph, and renders the result to
// stdout.
func listCatalogue(ctx context.Context, kind usecase.CatalogueKind, flags *catalogueFlags, args []string, stdout io.Writer) error {
	in := newListCatalogueInput(kind, flags, args)

	return runUseCase(ctx, func(runCtx context.Context, uc *usecase.ListCatalogueUseCase) error {
		result, err := uc.Execute(runCtx, in)
		if err != nil {
			return err
		}
		return renderCatalogue(stdout, result)
	})
}

// newListCatalogueInput maps the kind, the positional registry, and the parsed
// flags onto the use case's input.
func newListCatalogueInput(kind usecase.CatalogueKind, flags *catalogueFlags, args []string) usecase.ListCatalogueInput {
	return usecase.ListCatalogueInput{
		Kind:     kind,
		Registry: args[0],
		Search:   flags.Search,
		Sort:     flags.Sort,
		Order:    flags.Order,
		Page:     flags.paging.Page,
		Limit:    flags.paging.Limit,
	}
}

// renderCatalogue writes the catalogue table followed by the paging line.
func renderCatalogue(stdout io.Writer, result *usecase.ListCatalogueResult) error {
	headers, rows := catalogueTable(result)
	table := Table{Headers: headers, Rows: rows}
	if err := table.Render(stdout); err != nil {
		return err
	}

	_, err := fmt.Fprintln(stdout, CataloguePagingLine(result.Page, result.Limit, len(result.Entries)))
	return err
}

// catalogueTable builds the headers and rows for the result's kind: personas
// render NAME/MEMBERS, skills and agents render NAME/KIND.
func catalogueTable(result *usecase.ListCatalogueResult) ([]string, [][]string) {
	rows := make([][]string, len(result.Entries))
	if result.Kind == usecase.CataloguePersona {
		for i, entry := range result.Entries {
			rows[i] = []string{entry.Name, entry.Members}
		}
		return []string{colName, colMembers}, rows
	}

	for i, entry := range result.Entries {
		rows[i] = []string{entry.Name, string(result.Kind)}
	}
	return []string{colName, colKind}, rows
}
