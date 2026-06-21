package usecase

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/presentation"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DescribeRegistryUseCaseParams injects the collaborators the use case composes.
type DescribeRegistryUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Logger     *zap.Logger
}

// DescribeRegistryUseCase reads one registry and renders its full detail.
type DescribeRegistryUseCase struct {
	registries storage.RegistriesStore
	logger     *zap.Logger
}

// NewDescribeRegistryUseCase builds the use case from the injected collaborators.
func NewDescribeRegistryUseCase(params DescribeRegistryUseCaseParams) *DescribeRegistryUseCase {
	return &DescribeRegistryUseCase{
		registries: params.Registries,
		logger:     params.Logger,
	}
}

// describeFields is the ordered set --fields may select from for describe detail.
func (uc *DescribeRegistryUseCase) describeFields() []string {
	return []string{
		fieldName, fieldTransport, fieldURI, fieldRef,
		fieldAuth, fieldTLS, fieldSSHKey, fieldTimeout,
		fieldCreationTimestamp, fieldLastUpdatedTimestamp,
	}
}

// knownFields is the set --fields may select from.
func (uc *DescribeRegistryUseCase) knownFields() map[string]struct{} {
	known := make(map[string]struct{}, len(uc.describeFields()))
	for _, f := range uc.describeFields() {
		known[f] = struct{}{}
	}

	return known
}

// Execute runs the find → not-found → project → render pipeline, returning a
// *Error on the first failing step.
func (uc *DescribeRegistryUseCase) Execute(request *DescribeRegistryRequest) error {
	fields, err := uc.determineFields(request.Fields)
	if err != nil {
		return err
	}

	registry, err := uc.registries.FindByName(request.Context, request.Name)
	if err != nil {
		return NewIOError(fmt.Sprintf("read registry %q: %v", request.Name, err))
	}
	if registry == nil {
		return NewNotFoundError(fmt.Sprintf("registry %q does not exist", request.Name))
	}

	return uc.render(request, *registry, fields)
}

// determineFields validates the requested fields and forces name present and
// first; an empty request yields every field, so only populated ones render.
func (uc *DescribeRegistryUseCase) determineFields(requested []string) ([]string, error) {
	if len(requested) == 0 {
		return uc.describeFields(), nil
	}

	known := uc.knownFields()
	fields := []string{fieldName}
	seen := map[string]struct{}{fieldName: {}}
	for _, f := range requested {
		if _, ok := known[f]; !ok {
			return nil, NewUsageError(fmt.Sprintf("unknown field %q", f))
		}
		if _, dup := seen[f]; dup {
			continue
		}
		seen[f] = struct{}{}
		fields = append(fields, f)
	}

	return fields, nil
}

// render projects the registry onto the selected fields and writes the
// descriptor, logging the outcome.
func (uc *DescribeRegistryUseCase) render(request *DescribeRegistryRequest, registry types.Registry, fields []string) error {
	descriptor := presentation.Descriptor{Fields: uc.project(registry, fields)}

	uc.logger.Debug("registry described",
		zap.String(telemetry.FieldRegistryName, registry.Metadata.Name),
	)

	if err := descriptor.Render(request.Out()); err != nil {
		return NewIOError(fmt.Sprintf("render descriptor: %v", err))
	}

	return nil
}

// project maps the selected fields onto descriptor fields, skipping fields with
// no value so the default view shows only populated detail. The auth and tls
// blocks become nested sections; credential values are the stored env references,
// never resolved (FR-002).
func (uc *DescribeRegistryUseCase) project(registry types.Registry, fields []string) []presentation.Field {
	out := make([]presentation.Field, 0, len(fields))
	for _, name := range fields {
		if field, ok := uc.fieldFor(registry, name); ok {
			out = append(out, field)
		}
	}

	return out
}

// fieldFor builds the descriptor field for one selected field name, reporting
// false when the registry has no value for it. Leaf fields resolve through a
// value table; the two nested blocks (auth, tls) are sections.
func (uc *DescribeRegistryUseCase) fieldFor(registry types.Registry, name string) (presentation.Field, bool) {
	switch name {
	case fieldAuth:
		return uc.section(name, uc.authChildren(registry.Spec.Auth))
	case fieldTLS:
		return uc.section(name, uc.tlsChildren(registry.Spec.TLS))
	default:
		return uc.leaf(name, uc.leafValue(registry, name))
	}
}

// leafValue resolves the stored value of a leaf field; an unknown name yields the
// empty string, which leaf treats as absent.
func (uc *DescribeRegistryUseCase) leafValue(registry types.Registry, name string) string {
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
func (uc *DescribeRegistryUseCase) leaf(label, value string) (presentation.Field, bool) {
	if value == "" {
		return presentation.Field{}, false
	}

	return presentation.Field{Label: label, Value: value}, true
}

// section builds a section field, reporting false when it has no children.
func (uc *DescribeRegistryUseCase) section(label string, children []presentation.Field) (presentation.Field, bool) {
	if len(children) == 0 {
		return presentation.Field{}, false
	}

	return presentation.Field{Label: label, Children: children}, true
}

// authChildren renders the auth block as its stored env references (FR-002),
// omitting either credential that was not set.
func (uc *DescribeRegistryUseCase) authChildren(auth *types.Auth) []presentation.Field {
	if auth == nil {
		return nil
	}

	var children []presentation.Field
	if child, ok := uc.leaf("username", auth.Username); ok {
		children = append(children, child)
	}
	if child, ok := uc.leaf("password", auth.Password); ok {
		children = append(children, child)
	}

	return children
}

// tlsChildren renders the transport-security block, omitting unset settings.
func (uc *DescribeRegistryUseCase) tlsChildren(tls *types.TLS) []presentation.Field {
	if tls == nil {
		return nil
	}

	var children []presentation.Field
	if tls.SkipVerify {
		children = append(children, presentation.Field{Label: "skipVerify", Value: "true"})
	}
	if child, ok := uc.leaf("caCert", tls.CACert); ok {
		children = append(children, child)
	}
	if child, ok := uc.leaf("clientCert", tls.ClientCert); ok {
		children = append(children, child)
	}
	if child, ok := uc.leaf("clientKey", tls.ClientKey); ok {
		children = append(children, child)
	}

	return children
}

// DescribeRegistryRequest is the per-invocation input for describing a registry.
type DescribeRegistryRequest struct {
	context.Context
	out io.Writer

	Name   string
	Fields []string
}

// NewDescribeRegistryRequest builds a request bound to ctx and writing to out.
func NewDescribeRegistryRequest(ctx context.Context, out io.Writer) *DescribeRegistryRequest {
	return &DescribeRegistryRequest{Context: ctx, out: out}
}

// Out returns the writer the command's output goes to.
func (r *DescribeRegistryRequest) Out() io.Writer {
	return r.out
}
