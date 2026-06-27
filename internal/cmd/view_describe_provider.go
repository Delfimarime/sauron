package cmd

import (
	"fmt"
	"io"
	"sort"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

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
	if len(requested) == 0 {
		return describeProviderFieldOrder, nil
	}

	known := make(map[string]struct{}, len(describeProviderFieldOrder))
	for _, f := range describeProviderFieldOrder {
		known[f] = struct{}{}
	}

	fields := []string{describeProviderFieldName}
	seen := map[string]struct{}{describeProviderFieldName: {}}
	for _, f := range requested {
		if _, ok := known[f]; !ok {
			return nil, fmt.Errorf("%w: unknown field %q", errInvalidFlag, f)
		}
		if _, dup := seen[f]; dup {
			continue
		}
		seen[f] = struct{}{}
		fields = append(fields, f)
	}

	return fields, nil
}

// renderDescribeProvider projects the selected fields onto a descriptor and writes
// it, skipping fields the provider has no value for.
func renderDescribeProvider(w io.Writer, provider *types.Provider, fields []string) error {
	view := descriptor{Fields: projectProvider(*provider, fields)}
	if err := view.render(w); err != nil {
		return usecase.NewIOError(fmt.Sprintf("render descriptor: %v", err))
	}
	return nil
}

// renderNoProvider writes the none-set line; reporting it is a successful outcome.
func renderNoProvider(w io.Writer) error {
	if _, err := fmt.Fprintln(w, noProviderMessage); err != nil {
		return usecase.NewIOError(fmt.Sprintf("render provider: %v", err))
	}
	return nil
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
