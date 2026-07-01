package cmd

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DescribeRegistry builds the `registry` subcommand of `describe`.
func DescribeRegistry() *cobra.Command {
	var flags fieldsFlags
	return newCommand("registry", "Show the configured registry's full detail",
		withLong("Registry prints the configured registry's full detail as a vertical key-value view, with field selection."),
		withArgs(cobra.NoArgs),
		withFlags(func(cmd *cobra.Command) { bindFieldsFlags(cmd, &flags, "source") }),
		withRunE(func(ctx context.Context, _ []string, stdout io.Writer) error {
			return describeRegistry(ctx, &flags, stdout)
		}),
	)
}

// describeRegistry holds the cobra-free logic: it validates the requested fields
// at this boundary, lets the fx graph invoke the use case, and renders the
// returned detail, returning the classified failure to the caller. An unknown
// field yields a usage error before the use case runs.
func describeRegistry(ctx context.Context, flags *fieldsFlags, stdout io.Writer) error {
	fields, err := selectDescribeFields(flags.Fields)
	if err != nil {
		return err
	}

	registry, err := runUseCase(ctx, func(runCtx context.Context, uc usecase.UseCase[usecase.DescribeRegistryRequest, types.Registry]) (*types.Registry, error) {
		return uc.Execute(runCtx, usecase.DescribeRegistryRequest{})
	})
	if err != nil {
		return err
	}

	return renderDescribeRegistry(stdout, registry, fields)
}

// describe field names; this view owns the valid set --fields may select from.
const (
	describeFieldSource      = "source"
	describeFieldTransport   = "transport"
	describeFieldRevision    = "revision"
	describeFieldCredentials = "credentials"
	describeFieldTLS         = "tls"
	describeFieldSSHKey      = "sshKey"
	describeFieldTimeout     = "timeout"
	describeFieldCreated     = "createdAt"
	describeFieldUpdated     = "lastUpdatedAt"
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
	return selectFields(requested, describeFieldOrder, describeFieldSource)
}

// renderDescribeRegistry projects the selected fields onto a descriptor and
// writes it, skipping fields the registry has no value for.
func renderDescribeRegistry(w io.Writer, registry *types.Registry, fields []string) error {
	view := descriptor{Fields: projectRegistry(*registry, fields)}
	ew := newErrWriter(w)
	ew.record(view.render(w))
	return ew.toIOError("render descriptor")
}

// projectRegistry maps the selected fields onto descriptor fields, skipping
// fields with no value so the default view shows only populated detail. The
// credentials and tls blocks become nested sections; credential values are the
// stored env references, never resolved.
func projectRegistry(registry types.Registry, fields []string) []descriptorField {
	out := make([]descriptorField, 0, len(fields))
	for _, name := range fields {
		if field, ok := registryFieldFor(registry, name); ok {
			out = append(out, field)
		}
	}

	return out
}

// registryFieldFor builds the descriptor field for one selected field name,
// reporting false when the registry has no value for it.
func registryFieldFor(registry types.Registry, name string) (descriptorField, bool) {
	switch name {
	case describeFieldCredentials:
		return sectionField(name, credentialsChildren(registry.Spec.Credentials))
	case describeFieldTLS:
		return sectionField(name, tlsChildren(registry.Spec.TLS))
	default:
		return leafField(name, registryLeafValue(registry, name))
	}
}

// registryLeafValue resolves the stored value of a leaf field; an unknown name
// yields the empty string, which leafField treats as absent.
func registryLeafValue(registry types.Registry, name string) string {
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
