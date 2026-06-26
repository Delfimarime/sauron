package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

var (
	// envRefPattern matches a credential expressed as an environment reference,
	// capturing the variable name.
	envRefPattern = regexp.MustCompile(`^\$\{env:([A-Za-z_][A-Za-z0-9_]*)\}$`)

	// artifactRoots are the locations whose presence proves a source hosts at least
	// one artifact.
	artifactRoots = []string{rootSkills, rootAgents}
)

// SetRegistryUseCaseParams injects the adapters and collaborators the use case
// composes.
type SetRegistryUseCaseParams struct {
	fx.In
	Filesystem extension.Registry `name:"registry.filesystem"`
	Git        extension.Registry `name:"registry.git"`
	HTTP       extension.Registry `name:"registry.http"`
	Open       OpenRegistry
	Registries storage.RegistriesStore
	Logger     *zap.Logger
}

// SetRegistryUseCase configures the single registry: it validates a reachable
// source hosts at least one artifact, then persists it — replacing any registry
// already configured.
type SetRegistryUseCase struct {
	adapters   map[types.Transport]extension.Registry
	open       OpenRegistry
	registries storage.RegistriesStore
	logger     *zap.Logger
}

// NewSetRegistryUseCase builds the use case from the injected adapters and
// collaborators.
func NewSetRegistryUseCase(params SetRegistryUseCaseParams) *SetRegistryUseCase {
	return &SetRegistryUseCase{
		adapters: map[types.Transport]extension.Registry{
			types.TransportFilesystem: params.Filesystem,
			types.TransportGit:        params.Git,
			types.TransportHTTP:       params.HTTP,
		},
		open:       params.Open,
		registries: params.Registries,
		logger:     params.Logger,
	}
}

// Execute runs the ordered upsert pipeline, returning a *Error on the first
// failing step. The existing registry is left unchanged until validation
// succeeds; on success it is replaced and the outcome is returned.
func (uc *SetRegistryUseCase) Execute(ctx context.Context, in SetRegistryInput) (*SetRegistryResult, error) {
	if err := uc.validateCredentialFormat(in.Password); err != nil {
		return nil, err
	}

	adapter, transport, err := uc.selectAdapter(in.Transport)
	if err != nil {
		return nil, err
	}

	opts := in.referenceOptions(transport)
	if err := classifyAdapterErr(adapter.Validate(opts...)); err != nil {
		return nil, err
	}

	if err := uc.probe(ctx, in, transport); err != nil {
		return nil, err
	}

	return uc.persist(ctx, in, transport)
}

// validateCredentialFormat requires the password (the secret) to be an
// environment reference, so a secret value is never typed on the command line or
// persisted. The username is not secret and may be a literal or a reference.
func (uc *SetRegistryUseCase) validateCredentialFormat(password string) error {
	if password != "" && !envRefPattern.MatchString(password) {
		return NewUsageError("the password must be a ${env:VAR} reference")
	}

	return nil
}

// selectAdapter resolves the adapter for the requested transport.
func (uc *SetRegistryUseCase) selectAdapter(transport string) (extension.Registry, types.Transport, error) {
	t := types.Transport(transport)
	if adapter, ok := uc.adapters[t]; ok {
		return adapter, t, nil
	}

	return nil, "", NewUsageError(fmt.Sprintf("unknown transport %q", transport))
}

// probe opens the source through the shared open action — which resolves
// credential references, builds the options, and opens the transport — then
// confirms the source hosts at least one artifact.
func (uc *SetRegistryUseCase) probe(ctx context.Context, in SetRegistryInput, transport types.Transport) error {
	fs, err := uc.open.Execute(ctx, in.toRegistry(transport))
	if err != nil {
		return err
	}

	return uc.scanArtifacts(ctx, fs)
}

// scanArtifacts reports unreachable when no artifact root yields an entry.
func (uc *SetRegistryUseCase) scanArtifacts(ctx context.Context, fs source.FileSystem) error {
	for _, root := range artifactRoots {
		files, err := fs.List(ctx, root, source.WithLimit(1))
		if err != nil {
			return NewUnreachableError(fmt.Sprintf("list %q: %v", root, err))
		}
		if len(files) > 0 {
			return nil
		}
	}

	return NewUnreachableError("hosts no artifact")
}

// persist builds the registry document, stamps its audit timestamps with the
// current instant in UTC (equal on create), and stores it, replacing any
// registry already present. It returns the configured outcome.
func (uc *SetRegistryUseCase) persist(ctx context.Context, in SetRegistryInput, transport types.Transport) (*SetRegistryResult, error) {
	registry := in.toRegistry(transport)
	now := time.Now().UTC().Format(time.RFC3339)
	registry.Metadata.CreatedAt = now
	registry.Metadata.LastUpdatedAt = now
	if err := uc.registries.Set(ctx, registry); err != nil {
		return nil, NewIOError(fmt.Sprintf("persist registry: %v", err))
	}

	uc.logger.Info("registry set",
		zap.String(telemetry.FieldRegistryURI, in.URI),
		zap.String(telemetry.FieldRegistryTransport, string(transport)),
	)

	return &SetRegistryResult{URI: in.URI, Transport: transport}, nil
}

// classifyAdapterErr maps an adapter failure to a classified use-case error: a
// usage class is preserved, anything else becomes unreachable.
func classifyAdapterErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, api.ErrUsage) {
		return NewUsageError(err.Error())
	}

	return NewUnreachableError(err.Error())
}

// SetRegistryResult is the presentation-agnostic outcome of configuring the
// registry: the URI now in effect and the transport it is reached over.
type SetRegistryResult struct {
	URI       string
	Transport types.Transport
}

// SetRegistryInput is the per-invocation input for configuring the registry.
type SetRegistryInput struct {
	URI       string
	Transport string
	Ref       string
	Username  string
	Password  string
	SSHKey    string

	SkipTLSVerify bool
	CACert        string
	ClientCert    string
	ClientKey     string

	Timeout time.Duration
}

// referenceOptions builds the options used to validate the input, carrying the
// raw credential references untouched.
func (r *SetRegistryInput) referenceOptions(transport types.Transport) []extension.Option {
	opts := []extension.Option{extension.WithURI(r.URI)}

	if transport == types.TransportGit && r.Ref != "" {
		opts = append(opts, extension.WithRef(r.Ref))
	}
	if r.Timeout > 0 {
		opts = append(opts, extension.WithTimeout(r.Timeout))
	}
	if r.SSHKey != "" {
		opts = append(opts, extension.WithSSHKey(r.SSHKey))
	}
	if r.Username != "" || r.Password != "" {
		opts = append(opts, extension.WithBasicAuth(r.Username, r.Password))
	}

	return append(opts, r.tlsOptions()...)
}

// tlsOptions builds the transport-security options from the input.
func (r *SetRegistryInput) tlsOptions() []extension.Option {
	var opts []extension.Option

	if r.SkipTLSVerify {
		opts = append(opts, extension.WithSkipTLSVerify(true))
	}
	if r.CACert != "" {
		opts = append(opts, extension.WithCACert(r.CACert))
	}
	if r.ClientCert != "" || r.ClientKey != "" {
		opts = append(opts, extension.WithClientCert(r.ClientCert, r.ClientKey))
	}

	return opts
}

// toRegistry assembles the persisted document, storing credential references
// verbatim and never the resolved values. The single registry carries no
// user-given name; spec.source is its identity.
func (r *SetRegistryInput) toRegistry(transport types.Transport) types.Registry {
	spec := types.RegistrySpec{
		Transport:   transport,
		Source:      r.URI,
		SSHKey:      r.SSHKey,
		Credentials: r.toCredentials(),
		TLS:         r.toTLS(),
	}

	if transport == types.TransportGit {
		spec.Revision = r.Ref
	}
	if r.Timeout > 0 {
		spec.Timeout = r.Timeout.String()
	}

	return types.Registry{Spec: spec}
}

// toCredentials returns the credential references, or nil when none were supplied.
func (r *SetRegistryInput) toCredentials() *types.Credentials {
	if r.Username == "" && r.Password == "" {
		return nil
	}

	return &types.Credentials{Username: r.Username, Password: r.Password}
}

// toTLS returns the transport-security block, or nil when none was supplied.
func (r *SetRegistryInput) toTLS() *types.TLS {
	if !r.SkipTLSVerify && r.CACert == "" && r.ClientCert == "" && r.ClientKey == "" {
		return nil
	}

	return &types.TLS{
		SkipVerify: r.SkipTLSVerify,
		CACert:     r.CACert,
		ClientCert: r.ClientCert,
		ClientKey:  r.ClientKey,
	}
}
