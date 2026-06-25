package cmd

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// the selectable columns of a registry listing and describe detail.
const (
	fieldName                 = "name"
	fieldTransport            = "transport"
	fieldURI                  = "uri"
	fieldRef                  = "ref"
	fieldAuth                 = "auth"
	fieldTLS                  = "tls"
	fieldSSHKey               = "sshKey"
	fieldTimeout              = "timeout"
	fieldCreationTimestamp    = "creationTimestamp"
	fieldLastUpdatedTimestamp = "lastUpdatedTimestamp"
)

// the sort directions a listing accepts.
const (
	orderAsc  = "asc"
	orderDesc = "desc"
)

// the columns a registry listing may select, sort by, and show by default.
var (
	registryListColumns = map[string]struct{}{
		fieldName: {}, fieldTransport: {}, fieldURI: {}, fieldRef: {}, fieldTimeout: {},
	}
	registryListDefaults = []string{fieldName, fieldTransport, fieldURI}
	registrySortColumns  = map[string]struct{}{fieldName: {}, fieldTransport: {}}
)

// registryProjectors maps each listing column to the registry field it reads.
func registryProjectors() map[string]func(types.Registry) string {
	return map[string]func(types.Registry) string{
		fieldName:      func(r types.Registry) string { return r.Metadata.Name },
		fieldTransport: func(r types.Registry) string { return string(r.Spec.Transport) },
		fieldURI:       func(r types.Registry) string { return r.Spec.URI },
		fieldRef:       func(r types.Registry) string { return r.Spec.Ref },
		fieldTimeout:   func(r types.Registry) string { return r.Spec.Timeout },
	}
}

// RegistryListOptions are the view options for the registry listing: the search
// term, the sort field and direction, and the selected columns.
type RegistryListOptions struct {
	Search string
	Sort   string
	Order  string
	Fields []string
}

// Validate reports a sentinel error for an unknown field, sort field, or order.
func (o RegistryListOptions) Validate() error {
	if _, err := selectFields(o.Fields, registryListColumns, registryListDefaults); err != nil {
		return err
	}
	sort, order := defaultSortOrder(o.Sort, o.Order)
	if _, ok := registrySortColumns[sort]; !ok {
		return fmt.Errorf("%w %q", errUnknownSort, sort)
	}
	if !isValidOrder(order) {
		return fmt.Errorf("%w %q", errUnknownOrder, order)
	}

	return nil
}

// RenderRegistryList filters, sorts, projects, and writes the registry listing
// as a table. An empty listing renders nothing.
func RenderRegistryList(w io.Writer, registries []types.Registry, opts RegistryListOptions) error {
	fields, err := selectFields(opts.Fields, registryListColumns, registryListDefaults)
	if err != nil {
		return err
	}
	sortBy, order := defaultSortOrder(opts.Sort, opts.Order)

	kept := filterRegistries(registries, opts.Search)
	sortRegistries(kept, sortBy, order)

	table := Table{Headers: fields, Rows: projectRows(kept, fields, registryProjectors())}

	return table.Render(w)
}

// filterRegistries keeps the registries whose name contains the term,
// case-insensitively; an empty term keeps every registry.
func filterRegistries(registries []types.Registry, search string) []types.Registry {
	term := strings.ToLower(search)
	kept := make([]types.Registry, 0, len(registries))
	for _, r := range registries {
		if strings.Contains(strings.ToLower(r.Metadata.Name), term) {
			kept = append(kept, r)
		}
	}

	return kept
}

// sortRegistries sorts by the sort field and direction, always breaking ties on
// name ascending for a deterministic order.
func sortRegistries(registries []types.Registry, sortBy, order string) {
	key := registrySortKey(sortBy)
	slices.SortStableFunc(registries, func(a, b types.Registry) int {
		primary := strings.Compare(key(a), key(b))
		if order == orderDesc {
			primary = -primary
		}
		return cmp.Or(primary, strings.Compare(a.Metadata.Name, b.Metadata.Name))
	})
}

// registrySortKey maps a registry to its comparison key for the sort field.
func registrySortKey(sortBy string) func(types.Registry) string {
	if sortBy == fieldTransport {
		return func(r types.Registry) string { return string(r.Spec.Transport) }
	}
	return func(r types.Registry) string { return r.Metadata.Name }
}

// registryDetailFields is the ordered set of fields a describe view may show; an
// empty selection yields them all, so only populated ones render.
var registryDetailFields = []string{
	fieldName, fieldTransport, fieldURI, fieldRef,
	fieldAuth, fieldTLS, fieldSSHKey, fieldTimeout,
	fieldCreationTimestamp, fieldLastUpdatedTimestamp,
}

// registryDetailColumns is the set a describe --fields may select from.
var registryDetailColumns = func() map[string]struct{} {
	known := make(map[string]struct{}, len(registryDetailFields))
	for _, f := range registryDetailFields {
		known[f] = struct{}{}
	}
	return known
}()

// RegistryDetailOptions are the view options for a describe: the selected fields.
type RegistryDetailOptions struct {
	Fields []string
}

// Validate reports a sentinel error for an unknown field.
func (o RegistryDetailOptions) Validate() error {
	_, err := selectFields(o.Fields, registryDetailColumns, registryDetailFields)
	return err
}

// RenderRegistryDetail projects the registry onto the selected fields and writes
// the descriptor, skipping fields with no value so the default view shows only
// populated detail. The auth and tls blocks become nested sections; credential
// values are the stored env references, never resolved.
func RenderRegistryDetail(w io.Writer, registry types.Registry, opts RegistryDetailOptions) error {
	fields, err := selectFields(opts.Fields, registryDetailColumns, registryDetailFields)
	if err != nil {
		return err
	}

	descriptor := Descriptor{Fields: projectRegistryDetail(registry, fields)}

	return descriptor.Render(w)
}

// projectRegistryDetail maps the selected fields onto descriptor fields,
// skipping any with no value.
func projectRegistryDetail(registry types.Registry, fields []string) []Field {
	out := make([]Field, 0, len(fields))
	for _, name := range fields {
		if field, ok := detailField(registry, name); ok {
			out = append(out, field)
		}
	}

	return out
}

// detailField builds the descriptor field for one selected field name, reporting
// false when the registry has no value for it. Leaf fields resolve through a
// value table; the two nested blocks (auth, tls) are sections.
func detailField(registry types.Registry, name string) (Field, bool) {
	switch name {
	case fieldAuth:
		return section(name, authChildren(registry.Spec.Auth))
	case fieldTLS:
		return section(name, tlsChildren(registry.Spec.TLS))
	default:
		return leaf(name, leafValue(registry, name))
	}
}

// leafValue resolves the stored value of a leaf field; an unknown name yields the
// empty string, which leaf treats as absent.
func leafValue(registry types.Registry, name string) string {
	values := map[string]string{
		fieldName:                 registry.Metadata.Name,
		fieldTransport:            string(registry.Spec.Transport),
		fieldURI:                  registry.Spec.URI,
		fieldRef:                  registry.Spec.Ref,
		fieldSSHKey:               registry.Spec.SSHKey,
		fieldTimeout:              registry.Spec.Timeout,
		fieldCreationTimestamp:    registry.Metadata.CreationTimestamp,
		fieldLastUpdatedTimestamp: registry.Metadata.LastUpdatedTimestamp,
	}

	return values[name]
}

// leaf builds a leaf field, reporting false for an empty value.
func leaf(label, value string) (Field, bool) {
	if value == "" {
		return Field{}, false
	}

	return Field{Label: label, Value: value}, true
}

// section builds a section field, reporting false when it has no children.
func section(label string, children []Field) (Field, bool) {
	if len(children) == 0 {
		return Field{}, false
	}

	return Field{Label: label, Children: children}, true
}

// authChildren renders the auth block as its stored env references, omitting
// either credential that was not set.
func authChildren(auth *types.Auth) []Field {
	if auth == nil {
		return nil
	}

	var children []Field
	if child, ok := leaf("username", auth.Username); ok {
		children = append(children, child)
	}
	if child, ok := leaf("password", auth.Password); ok {
		children = append(children, child)
	}

	return children
}

// tlsChildren renders the transport-security block, omitting unset settings.
func tlsChildren(tls *types.TLS) []Field {
	if tls == nil {
		return nil
	}

	var children []Field
	if tls.SkipVerify {
		children = append(children, Field{Label: "skipVerify", Value: "true"})
	}
	if child, ok := leaf("caCert", tls.CACert); ok {
		children = append(children, child)
	}
	if child, ok := leaf("clientCert", tls.ClientCert); ok {
		children = append(children, child)
	}
	if child, ok := leaf("clientKey", tls.ClientKey); ok {
		children = append(children, child)
	}

	return children
}
