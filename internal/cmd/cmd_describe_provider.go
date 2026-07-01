package cmd

import (
	"context"
	"io"
	"sort"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DescribeProvider builds the `provider` subcommand of `describe`.
func DescribeProvider() *cobra.Command {
	var flags fieldsFlags
	return newCommand("provider", "Show the active provider's full detail",
		withLong("Provider prints the active provider's full detail as a vertical key-value view, with field selection."),
		withArgs(cobra.NoArgs),
		withFlags(func(cmd *cobra.Command) { bindFieldsFlags(cmd, &flags, "name") }),
		withRunE(func(ctx context.Context, _ []string, stdout io.Writer) error {
			return describeProvider(ctx, &flags, stdout)
		}),
	)
}

// describeProvider holds the cobra-free logic: it validates the requested fields
// at this boundary, lets the fx graph invoke the use case, and renders the
// returned detail. When no provider is set the use case returns nil; the handler
// prints the none-set line and exits successfully. An unknown field yields a usage
// error before the use case runs.
func describeProvider(ctx context.Context, flags *fieldsFlags, stdout io.Writer) error {
	fields, err := selectDescribeProviderFields(flags.Fields)
	if err != nil {
		return err
	}

	provider, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.DescribeProviderRequest, types.Provider]) (*types.Provider, error) {
		return uc.Execute(runCtx, usecase.DescribeProviderRequest{})
	})
	if err != nil {
		return err
	}
	if provider == nil {
		return renderNoProvider(stdout)
	}

	return renderDescribeProvider(stdout, provider, fields)
}

// describe-provider field names; this view owns the valid set --fields may select
// from. created/updated reuse the shared audit-field names.
const (
	describeProviderFieldName            = "name"
	describeProviderFieldDirectory       = "directory"
	describeProviderFieldLabels          = "labels"
	describeProviderFieldLastSynced      = "lastSyncedAt"
	describeProviderFieldLastSyncAttempt = "lastSyncAttemptAt"
)

// noProviderMessage is the line printed when no provider is set; reporting it is
// not an error and the command exits successfully.
const noProviderMessage = "no provider is set"

// describeProviderFieldOrder is the ordered set --fields may select from. name is
// the provider's identity and is always present and first.
var describeProviderFieldOrder = []string{
	describeProviderFieldName, describeProviderFieldDirectory, describeProviderFieldLabels,
	describeFieldCreated, describeFieldUpdated,
	describeProviderFieldLastSynced, describeProviderFieldLastSyncAttempt,
}

// selectDescribeProviderFields validates the requested fields against the field
// set, forcing name present and first and deduping; an empty request yields every
// field in order. An unknown field is a usage error (exit 2) raised before the use
// case runs.
func selectDescribeProviderFields(requested []string) ([]string, error) {
	return selectFields(requested, describeProviderFieldOrder, describeProviderFieldName)
}

// renderDescribeProvider projects the selected fields onto a descriptor and writes
// it, skipping fields the provider has no value for.
func renderDescribeProvider(w io.Writer, provider *types.Provider, fields []string) error {
	view := descriptor{Fields: projectProvider(*provider, fields)}
	ew := newErrWriter(w)
	ew.record(view.render(w))
	return ew.toIOError("render descriptor")
}

// renderNoProvider writes the none-set line; reporting it is a successful outcome.
func renderNoProvider(w io.Writer) error {
	ew := newErrWriter(w)
	ew.printf("%s\n", noProviderMessage)
	return ew.toIOError("render provider")
}

// projectProvider maps the selected fields onto descriptor fields, skipping fields
// with no value so the default view shows only populated detail. The labels block
// becomes a nested section with its keys sorted; directory is derived from the
// provider name.
func projectProvider(provider types.Provider, fields []string) []descriptorField {
	out := make([]descriptorField, 0, len(fields))
	for _, name := range fields {
		if field, ok := providerFieldFor(provider, name); ok {
			out = append(out, field)
		}
	}

	return out
}

// providerFieldFor builds the descriptor field for one selected field name,
// reporting false when the provider has no value for it.
func providerFieldFor(provider types.Provider, name string) (descriptorField, bool) {
	if name == describeProviderFieldLabels {
		return sectionField(name, labelChildren(provider.Metadata.Labels))
	}

	return leafField(name, providerLeafValue(provider, name))
}

// providerLeafValue resolves the stored value of a leaf field; an unknown name
// yields the empty string, which leafField treats as absent. directory is derived
// from the provider name, never stored.
func providerLeafValue(provider types.Provider, name string) string {
	values := map[string]string{
		describeProviderFieldName:            provider.Metadata.Name,
		describeProviderFieldDirectory:       providerDirectory(provider.Metadata.Name),
		describeFieldCreated:                 provider.Metadata.CreatedAt,
		describeFieldUpdated:                 provider.Metadata.LastUpdatedAt,
		describeProviderFieldLastSynced:      provider.Spec.LastSyncedAt,
		describeProviderFieldLastSyncAttempt: provider.Spec.LastSyncAttemptAt,
	}

	return values[name]
}

// providerDirectory derives the home-relative directory a provider installs into
// from its name; an unrecognized name yields the empty string, treated as absent.
func providerDirectory(name string) string {
	switch name {
	case types.ProviderClaude:
		return "~/.claude"
	case types.ProviderZencoder:
		return "~/.zencoder"
	default:
		return ""
	}
}

// labelChildren renders the labels block as descriptor fields with its keys
// sorted for deterministic output, returning nil when there are none.
func labelChildren(labels map[string]string) []descriptorField {
	if len(labels) == 0 {
		return nil
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	children := make([]descriptorField, 0, len(keys))
	for _, key := range keys {
		children = append(children, descriptorField{Label: key, Value: labels[key]})
	}

	return children
}
