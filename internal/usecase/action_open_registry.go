package usecase

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// OpenRegistryActionParams injects the named transport adapters and collaborators
// the action composes.
type OpenRegistryActionParams struct {
	fx.In
	Filesystem extension.Registry `name:"registry.filesystem"`
	Git        extension.Registry `name:"registry.git"`
	HTTP       extension.Registry `name:"registry.http"`
	Logger     *zap.Logger
}

// OpenRegistry opens a stored registry's source. It is the seam downstream use
// cases compose and tests mock.
type OpenRegistry interface {
	// Execute opens registry's source over its transport, returning the
	// read-only file system, or a classified *Error.
	Execute(ctx context.Context, registry types.Registry) (source.FileSystem, error)
}

// OpenRegistryAction opens a stored registry's source: it selects the transport
// adapter, resolves ${env:VAR} credential references, builds the extension option
// set from the registry spec, and opens the source. It is the shared
// open-a-stored-registry step the catalogue, install, and listing use cases
// compose. An open failure classifies as unreachable.
type OpenRegistryAction struct {
	logger    *zap.Logger
	lookupEnv func(string) (string, bool)
	adapters  map[types.Transport]extension.Registry
}

// NewOpenRegistryAction builds the action from the injected adapters and
// collaborators.
func NewOpenRegistryAction(params OpenRegistryActionParams) *OpenRegistryAction {
	return &OpenRegistryAction{
		adapters: map[types.Transport]extension.Registry{
			types.TransportFilesystem: params.Filesystem,
			types.TransportGit:        params.Git,
			types.TransportHTTP:       params.HTTP,
		},
		lookupEnv: os.LookupEnv,
		logger:    params.Logger,
	}
}

// Execute opens registry's source over its transport, returning the read-only
// file system. It returns a usage *Error for an unknown transport, an unreachable
// *Error for an unset credential reference or a failed open.
func (a *OpenRegistryAction) Execute(ctx context.Context, registry types.Registry) (source.FileSystem, error) {
	adapter, ok := a.adapters[registry.Spec.Transport]
	if !ok {
		return nil, NewUsageError(fmt.Sprintf("unknown transport %q", registry.Spec.Transport))
	}

	opts, err := a.connectOptions(registry.Spec)
	if err != nil {
		return nil, err
	}

	fs, err := adapter.Open(ctx, opts...)
	if err != nil {
		return nil, classifyAdapterErr(err)
	}

	return fs, nil
}

// connectOptions builds the extension option set from the spec, resolving any
// ${env:VAR} credential references to their values for connecting only.
func (a *OpenRegistryAction) connectOptions(spec types.RegistrySpec) ([]extension.Option, error) {
	opts := []extension.Option{extension.WithURI(spec.Source)}

	if spec.Transport == types.TransportGit && spec.Revision != "" {
		opts = append(opts, extension.WithRef(spec.Revision))
	}
	if timeout, err := a.timeout(spec.Timeout); err != nil {
		return nil, err
	} else if timeout > 0 {
		opts = append(opts, extension.WithTimeout(timeout))
	}
	if spec.SSHKey != "" {
		opts = append(opts, extension.WithSSHKey(spec.SSHKey))
	}

	credentialsOpt, err := a.credentialsOption(spec.Credentials)
	if err != nil {
		return nil, err
	}
	if credentialsOpt != nil {
		opts = append(opts, credentialsOpt)
	}

	return append(opts, a.tlsOptions(spec.TLS)...), nil
}

// timeout parses the spec's Go duration string; an empty value yields no bound.
func (a *OpenRegistryAction) timeout(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, NewUsageError(fmt.Sprintf("invalid timeout %q: %v", value, err))
	}

	return timeout, nil
}

// credentialsOption builds the basic-auth option, resolving credential
// references; it returns nil when no credentials were supplied.
func (a *OpenRegistryAction) credentialsOption(credentials *types.Credentials) (extension.Option, error) {
	if credentials == nil || (credentials.Username == "" && credentials.Password == "") {
		return nil, nil
	}

	username, err := a.resolveRef(credentials.Username)
	if err != nil {
		return nil, err
	}
	password, err := a.resolveRef(credentials.Password)
	if err != nil {
		return nil, err
	}

	return extension.WithBasicAuth(username, password), nil
}

// resolveRef resolves a ${env:VAR} reference to its value; a literal (or empty)
// value is returned unchanged. A referenced but unset variable is unreachable.
func (a *OpenRegistryAction) resolveRef(value string) (string, error) {
	match := envRefPattern.FindStringSubmatch(value)
	if match == nil {
		return value, nil
	}

	resolved, ok := a.lookupEnv(match[1])
	if !ok {
		return "", NewUnreachableError(fmt.Sprintf("environment variable %q is not set", match[1]))
	}

	return resolved, nil
}

// tlsOptions builds the transport-security options from the spec.
func (a *OpenRegistryAction) tlsOptions(tls *types.TLS) []extension.Option {
	if tls == nil {
		return nil
	}

	var opts []extension.Option
	if tls.SkipVerify {
		opts = append(opts, extension.WithSkipTLSVerify(true))
	}
	if tls.CACert != "" {
		opts = append(opts, extension.WithCACert(tls.CACert))
	}
	if tls.ClientCert != "" || tls.ClientKey != "" {
		opts = append(opts, extension.WithClientCert(tls.ClientCert, tls.ClientKey))
	}

	return opts
}
