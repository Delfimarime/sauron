package usecase

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// OpenRegistryUseCaseParams injects the named transport adapters and collaborators
// the action composes.
type OpenRegistryUseCaseParams struct {
	fx.In
	Git    extension.Registry `name:"registry.git"`
	HTTP   extension.Registry `name:"registry.http"`
	Logger *zap.Logger
}

// OpenRegistryUseCase opens a stored registry's source. It is the seam
// downstream use cases compose and tests mock.
type OpenRegistryUseCase interface {
	// Execute opens registry's source over its transport, returning the
	// read-only file system, or a classified *Error.
	Execute(ctx context.Context, registry types.Registry) (source.FileSystem, error)
}

// openRegistryUseCase opens a stored registry's source: it selects the transport
// adapter, resolves ${env:VAR} credential references, builds the extension option
// set from the registry spec, and opens the source. It is the shared
// open-a-stored-registry step the catalogue, install, and listing use cases
// compose. An open failure classifies as unreachable.
type openRegistryUseCase struct {
	logger    *zap.Logger
	lookupEnv func(string) (string, bool)
	adapters  map[types.Transport]extension.Registry
}

// NewOpenRegistryUseCase builds the use case from the injected adapters and
// collaborators.
func NewOpenRegistryUseCase(params OpenRegistryUseCaseParams) OpenRegistryUseCase {
	return &openRegistryUseCase{
		adapters: map[types.Transport]extension.Registry{
			types.TransportGit:  params.Git,
			types.TransportHTTP: params.HTTP,
		},
		lookupEnv: os.LookupEnv,
		logger:    params.Logger,
	}
}

// Execute opens registry's source over its transport, returning the read-only
// file system. It returns a usage *Error for an unknown transport, an unreachable
// *Error for an unset credential reference or a failed open.
func (a *openRegistryUseCase) Execute(ctx context.Context, registry types.Registry) (source.FileSystem, error) {
	adapter, ok := a.adapters[registry.Spec.Transport]
	if !ok {
		return nil, NewUsageError(fmt.Sprintf("unknown transport %q", registry.Spec.Transport))
	}

	opts, err := connectOptions(registry.Spec, a.resolveRef)
	if err != nil {
		return nil, err
	}

	fs, err := adapter.Open(ctx, opts...)
	if err != nil {
		return nil, classifyAdapterErr(err)
	}

	return fs, nil
}

// resolveRef resolves a ${env:VAR} reference to its value; a literal (or empty)
// value is returned unchanged. A referenced but unset variable is unreachable.
func (a *openRegistryUseCase) resolveRef(value string) (string, error) {
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
