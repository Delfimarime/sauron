package cmd

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// the installed-artifact table column headers beyond the shared tableHeaderName.
const (
	artifactFieldVersion       = "version"
	artifactFieldLastUpdatedAt = "lastUpdatedAt"
)

// artifactFieldOrder is the ordered set --fields may select from; name is the
// artifact's identity and is always present and first.
var artifactFieldOrder = []string{tableHeaderName, artifactFieldVersion, artifactFieldLastUpdatedAt}

// artifactSortFields are the fields an artifact listing may sort by — the set
// this view owns and validates --sort against.
var artifactSortFields = []string{tableHeaderName, artifactFieldLastUpdatedAt}

// ListSkills builds the `skills` subcommand of `list`.
func ListSkills() *cobra.Command {
	return newListArtifactsCommand(
		types.KindSkill, "skills",
		"List the installed skills",
		"Skills lists the installed skills as a name/version/lastUpdatedAt table (name only by default) followed by a paging line.",
	)
}

// ListAgents builds the `agents` subcommand of `list`.
func ListAgents() *cobra.Command {
	return newListArtifactsCommand(
		types.KindAgent, "agents",
		"List the installed agents",
		"Agents lists the installed agents as a name/version/lastUpdatedAt table (name only by default) followed by a paging line.",
	)
}

// newListArtifactsCommand builds one artifact-listing leaf command bound to
// its kind; the leaves differ only by the kind their run closure captures and
// passes to listInstalledArtifacts.
func newListArtifactsCommand(kind, use, short, long string) *cobra.Command {
	var flags listFlags
	var fields fieldsFlags
	return newCommand(use, short,
		withLong(long),
		withArgs(cobra.NoArgs),
		withFlags(func(cmd *cobra.Command) {
			bindListFlags(cmd, &flags, "artifact")
			bindFieldsFlags(cmd, &fields, tableHeaderName)
		}),
		withRunE(func(ctx context.Context, _ []string, stdout io.Writer) error {
			return listInstalledArtifacts(ctx, kind, &flags, &fields, stdout)
		}),
	)
}

// listInstalledArtifacts holds the cobra-free logic: it validates the
// requested fields and the sort/order at this boundary, before the use case
// runs, lets the fx graph invoke it, then renders the returned page as a
// table under the selected columns followed by the shared paging line.
func listInstalledArtifacts(ctx context.Context, kind string, flags *listFlags, fields *fieldsFlags, stdout io.Writer) error {
	selected, err := selectArtifactFields(fields.Fields)
	if err != nil {
		return err
	}

	in, err := newListArtifactsInput(kind, flags)
	if err != nil {
		return err
	}

	result, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.ListArtifactsRequest, usecase.ListArtifactsResponse]) (*usecase.ListArtifactsResponse, error) {
		return uc.Execute(runCtx, in)
	})
	if err != nil {
		return err
	}

	ew := newErrWriter(stdout)
	rows := buildTable(selected, result.Items, func(a types.Artifact) []string {
		return artifactRow(a, selected)
	})
	ew.record(rows.render(stdout))
	ew.printf("%s\n", pagingLine(result.Page, result.Limit, result.Offset, len(result.Items)))
	return ew.toIOError("render artifacts")
}

// newListArtifactsInput maps the kind and the parsed flags onto the use
// case's input, defaulting and validating the view options (--sort, --order)
// at this boundary; an invalid value yields a usage error before the use case
// runs. --search is a free substring and is not validated.
func newListArtifactsInput(kind string, flags *listFlags) (usecase.ListArtifactsRequest, error) {
	sortField := tableHeaderName
	if flags.Sort != "" {
		sortField = flags.Sort
	}
	if !slices.Contains(artifactSortFields, sortField) {
		return usecase.ListArtifactsRequest{}, fmt.Errorf("%w: unknown sort field %q", errInvalidFlag, sortField)
	}

	order := defaultOrder(flags.Order)
	if err := validateOrder(order); err != nil {
		return usecase.ListArtifactsRequest{}, err
	}

	return usecase.ListArtifactsRequest{
		Kind: kind,
		ListWindow: usecase.ListWindow{
			Search: flags.Search,
			Sort:   sortField,
			Order:  order,
			Page:   flags.paging.Page,
			Limit:  flags.paging.Limit,
		},
	}, nil
}

// selectArtifactFields resolves the columns an artifact listing renders: the
// default view (no --fields given) shows name only (FR-002), while a non-empty
// request selects and reorders columns through the shared selectFields
// validator, name forced present and first. This default differs from the
// shared selectFields contract (which yields every field for an empty
// request) because the artifact listing's default view is narrower than its
// full column set.
func selectArtifactFields(requested []string) ([]string, error) {
	if len(requested) == 0 {
		return []string{tableHeaderName}, nil
	}

	return selectFields(requested, artifactFieldOrder, tableHeaderName)
}

// artifactRow projects one artifact onto the selected, ordered column values.
func artifactRow(a types.Artifact, fields []string) []string {
	out := make([]string, len(fields))
	for i, field := range fields {
		out[i] = artifactFieldValue(a, field)
	}
	return out
}

// artifactFieldValue resolves the stored value of one selected column;
// selectFields already rejected any field outside artifactFieldOrder, so an
// unrecognized name here can only be the identity column, name.
func artifactFieldValue(a types.Artifact, field string) string {
	switch field {
	case artifactFieldVersion:
		return a.Spec.Version
	case artifactFieldLastUpdatedAt:
		return a.Spec.UpdatedAt
	default:
		return a.Metadata.Name
	}
}
