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
	// namePattern is the path-safe registry-name grammar enforced before any source
	// is contacted.
	namePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

	// envRefPattern matches a credential expressed as an environment reference,
	// capturing the variable name.
	envRefPattern = regexp.MustCompile(`^\$\{env:([A-Za-z_][A-Za-z0-9_]*)\}$`)

	// artifactRoots are the locations whose presence proves a source hosts at least
	// one artifact.
	artifactRoots = []string{rootSkills, rootAgents}
)

// AddRegistryUseCaseParams injects the adapters and collaborators the use case
// composes.
type AddRegistryUseCaseParams struct {
	fx.In
	Filesystem extension.Registry `name:"registry.filesystem"`
	Git        extension.Registry `name:"registry.git"`
	HTTP       extension.Registry `name:"registry.http"`
	Open       OpenRegistry
	Registries storage.RegistriesStore
	Logger     *zap.Logger
}

// AddRegistryUseCase registers a validated, reachable source.
type AddRegistryUseCase struct {
	adapters   map[types.Transport]extension.Registry
	open       OpenRegistry
	registries storage.RegistriesStore
	logger     *zap.Logger
}

// NewAddRegistryUseCase builds the use case from the injected adapters and
// collaborators.
func NewAddRegistryUseCase(params AddRegistryUseCaseParams) *AddRegistryUseCase {
	return &AddRegistryUseCase{
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

// Execute runs the ordered registration pipeline, returning the persisted
// registry, or a *Error on the first failing step.
func (uc *AddRegistryUseCase) Execute(ctx context.Context, in AddRegistryInput) (*types.Registry, error) {
	if err := uc.validateName(in.Name); err != nil {
		return nil, err
	}
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
	if err := uc.ensureUnique(ctx, in); err != nil {
		return nil, err
	}

	if err := uc.probe(ctx, in, transport); err != nil {
		return nil, err
	}

	return uc.persist(ctx, in, transport)
}

// validateName enforces the path-safe name grammar.
func (uc *AddRegistryUseCase) validateName(name string) error {
	if namePattern.MatchString(name) {
		return nil
	}

	return NewUsageError(fmt.Sprintf("name %q is not path-safe", name))
}

// validateCredentialFormat requires the password (the secret) to be an
// environment reference, so a secret value is never typed on the command line or
// persisted. The username is not secret and may be a literal or a reference.
func (uc *AddRegistryUseCase) validateCredentialFormat(password string) error {
	if password != "" && !envRefPattern.MatchString(password) {
		return NewUsageError("the password must be a ${env:VAR} reference")
	}

	return nil
}

// selectAdapter resolves the adapter for the requested transport.
func (uc *AddRegistryUseCase) selectAdapter(transport string) (extension.Registry, types.Transport, error) {
	t := types.Transport(transport)
	if adapter, ok := uc.adapters[t]; ok {
		return adapter, t, nil
	}

	return nil, "", NewUsageError(fmt.Sprintf("unknown transport %q", transport))
}

// ensureUnique rejects a name already taken by a stored registry.
func (uc *AddRegistryUseCase) ensureUnique(ctx context.Context, in AddRegistryInput) error {
	valueOf, err := uc.registries.FindByName(ctx, in.Name)
	if err != nil {
		return NewIOError(fmt.Sprintf("lookup registry %q: %v", in.Name, err))
	}
	if valueOf != nil {
		return NewConflictError(fmt.Sprintf("registry %q already exists", in.Name))
	}
	return nil
}

// probe opens the source through the shared open action — which resolves
// credential references, builds the options, and opens the transport — then
// confirms the source hosts at least one artifact.
func (uc *AddRegistryUseCase) probe(ctx context.Context, in AddRegistryInput, transport types.Transport) error {
	fs, err := uc.open.Execute(ctx, in.toRegistry(transport))
	if err != nil {
		return err
	}

	return uc.scanArtifacts(ctx, fs)
}

// scanArtifacts reports unreachable when no artifact root yields an entry.
func (uc *AddRegistryUseCase) scanArtifacts(ctx context.Context, fs source.FileSystem) error {
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
// current instant in UTC (equal on create), stores it, and returns the stamped
// document.
func (uc *AddRegistryUseCase) persist(ctx context.Context, in AddRegistryInput, transport types.Transport) (*types.Registry, error) {
	registry := in.toRegistry(transport)
	now := time.Now().UTC().Format(time.RFC3339)
	registry.Metadata.CreationTimestamp = now
	registry.Metadata.LastUpdatedTimestamp = now
	if err := uc.registries.Add(ctx, registry); err != nil {
		return nil, NewIOError(fmt.Sprintf("persist registry %q: %v", in.Name, err))
	}

	uc.logger.Info("registry registered",
		zap.String(telemetry.FieldRegistryName, in.Name),
		zap.String(telemetry.FieldRegistryTransport, string(transport)),
	)

	return &registry, nil
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

// AddRegistryInput is the input for registering a source.
type AddRegistryInput struct {
	Name      string
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
func (in AddRegistryInput) referenceOptions(transport types.Transport) []extension.Option {
	opts := []extension.Option{extension.WithURI(in.URI)}

	if transport == types.TransportGit && in.Ref != "" {
		opts = append(opts, extension.WithRef(in.Ref))
	}
	if in.Timeout > 0 {
		opts = append(opts, extension.WithTimeout(in.Timeout))
	}
	if in.SSHKey != "" {
		opts = append(opts, extension.WithSSHKey(in.SSHKey))
	}
	if in.Username != "" || in.Password != "" {
		opts = append(opts, extension.WithBasicAuth(in.Username, in.Password))
	}

	return append(opts, in.tlsOptions()...)
}

// tlsOptions builds the transport-security options from the input.
func (in AddRegistryInput) tlsOptions() []extension.Option {
	var opts []extension.Option

	if in.SkipTLSVerify {
		opts = append(opts, extension.WithSkipTLSVerify(true))
	}
	if in.CACert != "" {
		opts = append(opts, extension.WithCACert(in.CACert))
	}
	if in.ClientCert != "" || in.ClientKey != "" {
		opts = append(opts, extension.WithClientCert(in.ClientCert, in.ClientKey))
	}

	return opts
}

// toRegistry assembles the persisted document, storing credential references
// verbatim and never the resolved values.
func (in AddRegistryInput) toRegistry(transport types.Transport) types.Registry {
	spec := types.RegistrySpec{
		Transport: transport,
		URI:       in.URI,
		SSHKey:    in.SSHKey,
		Auth:      in.toAuth(),
		TLS:       in.toTLS(),
	}

	if transport == types.TransportGit {
		spec.Ref = in.Ref
	}
	if in.Timeout > 0 {
		spec.Timeout = in.Timeout.String()
	}

	return types.Registry{
		Metadata: types.Metadata{Name: in.Name},
		Spec:     spec,
	}
}

// toAuth returns the credential references, or nil when none were supplied.
func (in AddRegistryInput) toAuth() *types.Auth {
	if in.Username == "" && in.Password == "" {
		return nil
	}

	return &types.Auth{Username: in.Username, Password: in.Password}
}

// toTLS returns the transport-security block, or nil when none was supplied.
func (in AddRegistryInput) toTLS() *types.TLS {
	if !in.SkipTLSVerify && in.CACert == "" && in.ClientCert == "" && in.ClientKey == "" {
		return nil
	}

	return &types.TLS{
		SkipVerify: in.SkipTLSVerify,
		CACert:     in.CACert,
		ClientCert: in.ClientCert,
		ClientKey:  in.ClientKey,
	}
}
