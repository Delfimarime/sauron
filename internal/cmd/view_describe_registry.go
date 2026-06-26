package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// describe field names; this view owns the valid set --fields may select from.
const (
	describeFieldSource      = "source"
	describeFieldTransport   = "transport"
	describeFieldRevision    = "revision"
	describeFieldCredentials = "credentials"
	describeFieldTLS         = "tls"
	describeFieldSSHKey      = "sshKey"
	describeFieldTimeout     = "timeout"
	describeFieldCreated     = "created"
	describeFieldUpdated     = "updated"
)

// describeFieldOrder is the ordered set --fields may select from. The single
// registry has no name; source is its identity and is always present and first.
var describeFieldOrder = []string{
	describeFieldSource, describeFieldTransport, describeFieldRevision,
	describeFieldCredentials, describeFieldTLS, describeFieldSSHKey, describeFieldTimeout,
	describeFieldCreated, describeFieldUpdated,
}

// selectDescribeFields validates the requested fields against the describe field
// set, forcing source present and first and deduping; an empty request yields
// every field in order. An unknown field is a usage error (exit 2) raised before
// the use case runs.
func selectDescribeFields(requested []string) ([]string, error) {
	if len(requested) == 0 {
		return describeFieldOrder, nil
	}

	known := make(map[string]struct{}, len(describeFieldOrder))
	for _, f := range describeFieldOrder {
		known[f] = struct{}{}
	}

	fields := []string{describeFieldSource}
	seen := map[string]struct{}{describeFieldSource: {}}
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

// renderDescribeRegistry projects the selected fields onto a descriptor and
// writes it, skipping fields the registry has no value for.
func renderDescribeRegistry(w io.Writer, registry *types.Registry, fields []string) error {
	view := descriptor{Fields: projectRegistry(*registry, fields)}
	if err := view.render(w); err != nil {
		return usecase.NewIOError(fmt.Sprintf("render descriptor: %v", err))
	}
	return nil
}

// projectRegistry maps the selected fields onto descriptor fields, skipping
// fields with no value so the default view shows only populated detail. The
// credentials and tls blocks become nested sections; credential values are the
// stored env references, never resolved.
func projectRegistry(registry types.Registry, fields []string) []descriptorField {
	out := make([]descriptorField, 0, len(fields))
	for _, name := range fields {
		if field, ok := fieldFor(registry, name); ok {
			out = append(out, field)
		}
	}

	return out
}

// fieldFor builds the descriptor field for one selected field name, reporting
// false when the registry has no value for it.
func fieldFor(registry types.Registry, name string) (descriptorField, bool) {
	switch name {
	case describeFieldCredentials:
		return sectionField(name, credentialsChildren(registry.Spec.Credentials))
	case describeFieldTLS:
		return sectionField(name, tlsChildren(registry.Spec.TLS))
	default:
		return leafField(name, leafValue(registry, name))
	}
}

// leafValue resolves the stored value of a leaf field; an unknown name yields the
// empty string, which leafField treats as absent.
func leafValue(registry types.Registry, name string) string {
	values := map[string]string{
		describeFieldTransport: string(registry.Spec.Transport),
		describeFieldSource:    registry.Spec.Source,
		describeFieldRevision:  registry.Spec.Revision,
		describeFieldSSHKey:    registry.Spec.SSHKey,
		describeFieldTimeout:   registry.Spec.Timeout,
		describeFieldCreated:   registry.Metadata.CreatedAt,
		describeFieldUpdated:   registry.Metadata.LastUpdatedAt,
	}

	return values[name]
}

// leafField builds a leaf field, reporting false for an empty value.
func leafField(label, value string) (descriptorField, bool) {
	if value == "" {
		return descriptorField{}, false
	}

	return descriptorField{Label: label, Value: value}, true
}

// sectionField builds a section field, reporting false when it has no children.
func sectionField(label string, children []descriptorField) (descriptorField, bool) {
	if len(children) == 0 {
		return descriptorField{}, false
	}

	return descriptorField{Label: label, Children: children}, true
}

// credentialsChildren renders the credentials block as its stored env
// references, omitting either credential that was not set.
func credentialsChildren(credentials *types.Credentials) []descriptorField {
	if credentials == nil {
		return nil
	}

	var children []descriptorField
	if child, ok := leafField("username", credentials.Username); ok {
		children = append(children, child)
	}
	if child, ok := leafField("password", credentials.Password); ok {
		children = append(children, child)
	}

	return children
}

// tlsChildren renders the transport-security block, omitting unset settings.
func tlsChildren(tls *types.TLS) []descriptorField {
	if tls == nil {
		return nil
	}

	var children []descriptorField
	if tls.SkipVerify {
		children = append(children, descriptorField{Label: "skipVerify", Value: "true"})
	}
	if child, ok := leafField("caCert", tls.CACert); ok {
		children = append(children, child)
	}
	if child, ok := leafField("clientCert", tls.ClientCert); ok {
		children = append(children, child)
	}
	if child, ok := leafField("clientKey", tls.ClientKey); ok {
		children = append(children, child)
	}

	return children
}
