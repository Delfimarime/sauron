package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
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

// Execute runs the ordered registration pipeline, returning a *Error on the
// first failing step.
func (uc *AddRegistryUseCase) Execute(request *AddRegistryRequest) error {
	if err := uc.validateName(request.Name); err != nil {
		return err
	}
	if err := uc.validateCredentialFormat(request.Password); err != nil {
		return err
	}

	adapter, transport, err := uc.selectAdapter(request.Transport)
	if err != nil {
		return err
	}

	opts := request.referenceOptions(transport)
	if err := classifyAdapterErr(adapter.Validate(opts...)); err != nil {
		return err
	}
	if err := uc.ensureUnique(request); err != nil {
		return err
	}

	if err := uc.probe(request, transport); err != nil {
		return err
	}

	return uc.persist(request, transport)
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
func (uc *AddRegistryUseCase) ensureUnique(request *AddRegistryRequest) error {
	valueOf, err := uc.registries.FindByName(request.Context, request.Name)
	if err != nil {
		return NewIOError(fmt.Sprintf("lookup registry %q: %v", request.Name, err))
	}
	if valueOf != nil {
		return NewConflictError(fmt.Sprintf("registry %q already exists", request.Name))
	}
	return nil
}

// probe opens the source through the shared open action — which resolves
// credential references, builds the options, and opens the transport — then
// confirms the source hosts at least one artifact.
func (uc *AddRegistryUseCase) probe(request *AddRegistryRequest, transport types.Transport) error {
	fs, err := uc.open.Execute(request.Context, request.toRegistry(transport))
	if err != nil {
		return err
	}

	return uc.scanArtifacts(request.Context, fs)
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
// current instant in UTC (equal on create), and stores it.
func (uc *AddRegistryUseCase) persist(request *AddRegistryRequest, transport types.Transport) error {
	registry := request.toRegistry(transport)
	now := time.Now().UTC().Format(time.RFC3339)
	registry.Metadata.CreationTimestamp = now
	registry.Metadata.LastUpdatedTimestamp = now
	if err := uc.registries.Add(request.Context, registry); err != nil {
		return NewIOError(fmt.Sprintf("persist registry %q: %v", request.Name, err))
	}

	uc.logger.Info("registry registered",
		zap.String(telemetry.FieldRegistryName, request.Name),
		zap.String(telemetry.FieldRegistryTransport, string(transport)),
	)
	if _, err := fmt.Fprintf(request.Out(), "registered registry %q (%s)\n", request.Name, transport); err != nil {
		return NewIOError(fmt.Sprintf("write report: %v", err))
	}

	return nil
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

// AddRegistryRequest is the per-invocation input for registering a source.
type AddRegistryRequest struct {
	baseRequest

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

// NewAddRegistryRequest builds a request bound to ctx and writing to out.
func NewAddRegistryRequest(ctx context.Context, out io.Writer) *AddRegistryRequest {
	return &AddRegistryRequest{baseRequest: baseRequest{Context: ctx, out: out}}
}

// referenceOptions builds the options used to validate the request, carrying the
// raw credential references untouched.
func (r *AddRegistryRequest) referenceOptions(transport types.Transport) []extension.Option {
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

// tlsOptions builds the transport-security options from the request.
func (r *AddRegistryRequest) tlsOptions() []extension.Option {
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
// verbatim and never the resolved values.
func (r *AddRegistryRequest) toRegistry(transport types.Transport) types.Registry {
	spec := types.RegistrySpec{
		Transport: transport,
		URI:       r.URI,
		SSHKey:    r.SSHKey,
		Auth:      r.toAuth(),
		TLS:       r.toTLS(),
	}

	if transport == types.TransportGit {
		spec.Ref = r.Ref
	}
	if r.Timeout > 0 {
		spec.Timeout = r.Timeout.String()
	}

	return types.Registry{
		Metadata: types.Metadata{Name: r.Name},
		Spec:     spec,
	}
}

// toAuth returns the credential references, or nil when none were supplied.
func (r *AddRegistryRequest) toAuth() *types.Auth {
	if r.Username == "" && r.Password == "" {
		return nil
	}

	return &types.Auth{Username: r.Username, Password: r.Password}
}

// toTLS returns the transport-security block, or nil when none was supplied.
func (r *AddRegistryRequest) toTLS() *types.TLS {
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
