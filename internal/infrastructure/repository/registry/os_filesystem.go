package registry

import (
	"context"
	"fmt"

	"github.com/spf13/afero"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// osFactory opens a local-directory source on the host filesystem.
type osFactory struct{}

// newOSFactory builds an osFactory.
func newOSFactory() *osFactory {
	return &osFactory{}
}

// Validate rejects options the local-directory transport does not accept:
// credentials, transport security, an SSH key, or a reference.
func (osFactory) Validate(opts ...extension.Option) error {
	options := api.Resolve(opts)

	switch {
	case options.Ref != "":
		return fmt.Errorf("%w: a reference is not supported", api.ErrUsage)
	case api.HasAuth(options):
		return fmt.Errorf("%w: credentials are not supported", api.ErrUsage)
	case api.HasTLS(options):
		return fmt.Errorf("%w: transport security is not supported", api.ErrUsage)
	case options.SSHKey != "":
		return fmt.Errorf("%w: an SSH key is not supported", api.ErrUsage)
	default:
		return nil
	}
}

// Open verifies the URI names a readable directory and returns a read-only view
// rooted at it.
func (f osFactory) Open(_ context.Context, opts ...extension.Option) (source.FileSystem, error) {
	options := api.Resolve(opts)

	if err := f.Validate(opts...); err != nil {
		return nil, err
	}

	osFs := afero.NewOsFs()
	info, err := osFs.Stat(options.URI)
	if err != nil {
		return nil, fmt.Errorf("%w: open %q: %w", api.ErrRuntime, options.URI, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %q is not a directory", api.ErrRuntime, options.URI)
	}

	return api.NewDirectory(afero.NewBasePathFs(osFs, options.URI)), nil
}
